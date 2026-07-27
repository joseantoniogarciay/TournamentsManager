// Package http expone adaptadores HTTP de la API.
package http

import (
	"net/http"
)

// NewHandler construye las rutas de infraestructura disponibles antes de los
// endpoints de negocio.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	return mux
}

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
