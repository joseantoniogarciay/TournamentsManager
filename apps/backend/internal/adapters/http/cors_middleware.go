package http

import (
	"net/http"
	"strings"
)

var corsAllowedMethods = map[string]struct{}{
	http.MethodDelete: {},
	http.MethodGet:    {},
	http.MethodPost:   {},
	http.MethodPut:    {},
}

var corsAllowedHeaders = map[string]struct{}{
	"authorization": {},
	"content-type":  {},
}

func requireAllowedOrigin(origins []string, next http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowedOrigins[origin] = struct{}{}
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}
		if _, allowed := allowedOrigins[origin]; !allowed {
			writeProblem(writer, http.StatusForbidden, "Origen no permitido")
			return
		}

		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Vary", "Origin")
		if request.Method != http.MethodOptions {
			next.ServeHTTP(writer, request)
			return
		}
		if !validPreflight(request) {
			writeProblem(writer, http.StatusForbidden, "Preflight no permitido")
			return
		}

		writer.Header().Set("Access-Control-Allow-Methods", "DELETE, GET, POST, PUT")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		writer.Header().Set("Access-Control-Max-Age", "600")
		writer.WriteHeader(http.StatusNoContent)
	})
}

func validPreflight(request *http.Request) bool {
	if _, allowed := corsAllowedMethods[request.Header.Get("Access-Control-Request-Method")]; !allowed {
		return false
	}
	for _, header := range strings.Split(request.Header.Get("Access-Control-Request-Headers"), ",") {
		header = strings.TrimSpace(strings.ToLower(header))
		if header == "" {
			continue
		}
		if _, allowed := corsAllowedHeaders[header]; !allowed {
			return false
		}
	}
	return true
}
