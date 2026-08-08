package http

import "net/http"

func requireCookieCSRF(trustedOrigins []string, next http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	for _, origin := range trustedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			panic("el origen CORS validado debe ser válido para la protección CSRF")
		}
	}
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
