package docs

import (
	_ "embed"
	"net/http"
)

//go:embed index.html
var scalarHTML []byte

//go:embed openapi.json
var openAPIJSON []byte

func Scalar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(scalarHTML)
}

func OpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(openAPIJSON)
}
