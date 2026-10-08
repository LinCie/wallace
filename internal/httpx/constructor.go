package httpx

import "net/http"

// Constructor wraps an HTTP handler with middleware.
type Constructor func(http.Handler) http.Handler
