package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var bodyValidator = newBodyValidator()

func newBodyValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return validate
}

// bodyError provides a friendly message while retaining the original error.
type bodyError struct {
	message string
	cause   error
}

func (e *bodyError) Error() string { return e.message }
func (e *bodyError) Unwrap() error { return e.cause }

// BindBody decodes a single JSON request body into dst, then validates its
// validate tags. dst must be a non-nil pointer to a struct. Fields follow
// encoding/json's binding rules. Input errors have user-friendly messages
// and wrap their underlying cause. Validation errors report the first failing
// field using its json tag, defaulting to its Go name.
func BindBody(r *http.Request, dst any) error {
	value := reflect.ValueOf(dst)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("bind body: destination must be a non-nil pointer to a struct")
	}
	if r.Body == nil {
		return &bodyError{message: "request body is required", cause: io.EOF}
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		message := "request body is malformed"
		var typeError *json.UnmarshalTypeError
		switch {
		case errors.Is(err, io.EOF):
			message = "request body is required"
		case errors.As(err, &typeError):
			name := typeError.Field
			if name == "" {
				name = "request body"
			}
			message = fmt.Sprintf("%s must be of type %s", name, typeError.Type)
		}
		return &bodyError{message: message, cause: err}
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("multiple JSON values in request body")
		}
		return &bodyError{message: "request body must contain a single JSON value", cause: err}
	}

	err := bodyValidator.Struct(dst)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return &bodyError{
			message: bodyValidationMessage(validationErrors[0]),
			cause:   err,
		}
	}
	return err
}

func bodyValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
