package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var queryValidator = newQueryValidator()

func newQueryValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := field.Tag.Get("query")
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return validate
}

// queryError provides a friendly message while retaining the original error.
type queryError struct {
	message string
	cause   error
}

func (e *queryError) Error() string { return e.message }
func (e *queryError) Unwrap() error { return e.cause }

// BindQuery binds query parameters to dst, then validates its validate tags.
// dst must be a non-nil pointer to a struct. Fields use query tags, defaulting
// to their Go names; query:"-" skips binding. Missing parameters leave fields
// unchanged. Scalars use the first value and slices use all repeated values.
// Supported types are strings, booleans, numbers, and pointers or slices of them.
// Input errors have user-friendly messages and wrap their underlying cause.
// Validation errors report the first failing field.
func BindQuery(r *http.Request, dst any) error {
	value := reflect.ValueOf(dst)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("bind query: destination must be a non-nil pointer to a struct")
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return &queryError{message: "query parameters are malformed", cause: err}
	}

	value = value.Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanSet() {
			continue
		}
		definition := value.Type().Field(i)
		name := definition.Tag.Get("query")
		if name == "-" {
			continue
		}
		if name == "" {
			name = definition.Name
		}
		values, ok := params[name]
		if !ok {
			continue
		}
		if err := bindQueryValue(field, values); err != nil {
			var conversionError *strconv.NumError
			if errors.As(err, &conversionError) {
				expected := "a number"
				switch conversionError.Func {
				case "ParseInt", "ParseUint":
					expected = "an integer"
				case "ParseBool":
					expected = "a boolean"
				}
				message := fmt.Sprintf("%s must be %s", name, expected)
				if errors.Is(err, strconv.ErrRange) {
					message = fmt.Sprintf("%s is out of range", name)
				}
				return &queryError{message: message, cause: err}
			}
			return fmt.Errorf("bind query %q: %w", name, err)
		}
	}

	err = queryValidator.Struct(dst)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return &queryError{
			message: queryValidationMessage(validationErrors[0]),
			cause:   err,
		}
	}
	return err
}

func queryValidationMessage(err validator.FieldError) string {
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

func bindQueryValue(field reflect.Value, values []string) error {
	switch field.Kind() {
	case reflect.Pointer:
		value := reflect.New(field.Type().Elem())
		if err := bindQueryValue(value.Elem(), values); err != nil {
			return err
		}
		field.Set(value)
	case reflect.Slice:
		value := reflect.MakeSlice(field.Type(), len(values), len(values))
		for i, raw := range values {
			if err := bindQueryValue(value.Index(i), []string{raw}); err != nil {
				return err
			}
		}
		field.Set(value)
	case reflect.String:
		field.SetString(values[0])
	case reflect.Bool:
		value, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		field.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(values[0], 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value, err := strconv.ParseUint(values[0], 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(values[0], field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(value)
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}
	return nil
}
