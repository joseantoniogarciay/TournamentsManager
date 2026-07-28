package http

import "net/http"

func requireCookieCSRF(next http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writeProblem(writer, http.StatusForbidden, "Origen no permitido")
	}))
	protected := protection.Handler(next)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		transport, ok := currentSessionTransport(request.Context())
		if ok && transport == cookieSession {
			protected.ServeHTTP(writer, request)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
