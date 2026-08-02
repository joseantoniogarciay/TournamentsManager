package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
)

type sessionAuthenticator interface {
	Authenticate(context.Context, string) (string, error)
}

type accountContextKey struct{}
type sessionTransportContextKey struct{}
type sessionTokenContextKey struct{}
type sessionCookieNameContextKey struct{}

type sessionTransport string

const (
	cookieSession sessionTransport = "cookie"
	bearerSession sessionTransport = "bearer"
)

type credential struct {
	token     string
	transport sessionTransport
}

func requireSession(authenticator sessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			credential, ok := sessionToken(request)
			if !ok {
				writeProblem(writer, http.StatusUnauthorized, "Sesión no válida")
				return
			}
			accountID, err := authenticator.Authenticate(request.Context(), credential.token)
			if err != nil {
				if errors.Is(err, leagues.ErrUnauthenticated) {
					writeProblem(writer, http.StatusUnauthorized, "Sesión no válida")
					return
				}
				writeProblem(writer, http.StatusInternalServerError, "No se pudo validar la sesión")
				return
			}
			requestContext := context.WithValue(request.Context(), accountContextKey{}, accountID)
			requestContext = context.WithValue(requestContext, sessionTransportContextKey{}, credential.transport)
			requestContext = context.WithValue(requestContext, sessionTokenContextKey{}, credential.token)
			next.ServeHTTP(writer, request.WithContext(requestContext))
		})
	}
}

func currentSessionToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(sessionTokenContextKey{}).(string)
	return token, ok && token != ""
}

func currentSessionTransport(ctx context.Context) (sessionTransport, bool) {
	transport, ok := ctx.Value(sessionTransportContextKey{}).(sessionTransport)
	return transport, ok
}

func currentAccountID(ctx context.Context) (string, bool) {
	accountID, ok := ctx.Value(accountContextKey{}).(string)
	return accountID, ok && accountID != ""
}

func sessionToken(request *http.Request) (credential, bool) {
	authorization := request.Header.Get("Authorization")
	cookie, cookieErr := request.Cookie(sessionCookieName(request.Context()))
	if authorization != "" && cookieErr == nil {
		return credential{}, false
	}
	if authorization != "" {
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authorization, bearerPrefix) || len(authorization) == len(bearerPrefix) {
			return credential{}, false
		}
		return credential{token: strings.TrimPrefix(authorization, bearerPrefix), transport: bearerSession}, true
	}
	if cookieErr == nil && cookie.Value != "" {
		return credential{token: cookie.Value, transport: cookieSession}, true
	}
	return credential{}, false
}

func sessionCookieName(ctx context.Context) string {
	name, ok := ctx.Value(sessionCookieNameContextKey{}).(string)
	if !ok || name == "" {
		return "__Host-tm_session"
	}
	return name
}
