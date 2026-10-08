package middleware

import (
	"log"
	"net/http"

	"wallace/internal/httpx"
)

// Recovery catches panics from downstream handlers and returns an internal
// server error instead of terminating the server.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				httpx.RespondError(
					w,
					http.StatusInternalServerError,
					"internal_error",
					"internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
