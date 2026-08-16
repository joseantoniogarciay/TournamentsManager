package http

import (
	"context"
	"net/http"
)

func requireCookieCSRF(trustedOrigins []string, next http.Handler) http.Handler {
	protection := http.NewCrossOriginProtection()
	for _, origin := range trustedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			panic("el origen CORS validado debe ser válido para la protección CSRF")
		}
	}
	protection.SetDenyHandler(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writeProblem(writer, http.StatusForbidden, "Origin is not allowed")
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

// refreshCookieCSRF applies origin checks only when refresh is transported by cookie.
func refreshCookieCSRF(trustedOrigins []string, cookies sessionCookieSettings, next http.Handler) http.Handler {
	protected := requireCookieCSRF(trustedOrigins, next)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if cookie, err := request.Cookie(cookies.refreshName); err == nil && cookie.Value != "" {
			ctx := context.WithValue(request.Context(), sessionTransportContextKey{}, cookieSession)
			protected.ServeHTTP(writer, request.WithContext(ctx))
			return
		}
		next.ServeHTTP(writer, request)
	})
}
