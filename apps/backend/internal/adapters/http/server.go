// Package http expone adaptadores HTTP de la API.
package http

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const (
	usernameAvailabilityLimit  = 30
	usernameAvailabilityWindow = time.Minute
)

// NewHandler construye las rutas de infraestructura disponibles antes de los
// endpoints de negocio.
func NewHandler(registrationService registration.Service, authenticator sessionAuthenticator, leagueService leagues.Service, corsAllowedOrigins []string) http.Handler {
	mux := http.NewServeMux()
	availabilityLimiter := newRequestLimiter(usernameAvailabilityLimit, usernameAvailabilityWindow)
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /v1/usernames/{username}/availability", usernameAvailability(registrationService, availabilityLimiter))
	mux.HandleFunc("POST /v1/registrations", register(registrationService))
	mux.HandleFunc("POST /v1/registration-verifications", verifyRegistration(registrationService))
	mux.HandleFunc("POST /v1/sessions/refresh", refreshSession(registrationService))
	mux.Handle("GET /v1/me/leagues", requireSession(authenticator)(http.HandlerFunc(listAccountLeagues(leagueService))))
	followHandler := requireSession(authenticator)(requireCookieCSRF(http.HandlerFunc(followLeague(leagueService))))
	mux.Handle("PUT /v1/me/leagues/{leagueId}/follow", followHandler)
	mux.Handle("DELETE /v1/me/leagues/{leagueId}/follow", followHandler)
	return requireAllowedOrigin(corsAllowedOrigins, mux)
}

func refreshSession(service registration.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") || len(authorization) == len("Bearer ") {
			writeProblem(writer, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		session, accessToken, refreshToken, err := service.Refresh(request.Context(), strings.TrimPrefix(authorization, "Bearer "))
		if errors.Is(err, registration.ErrRefreshInvalid) {
			writeProblem(writer, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo renovar la sesión")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": "bearer", "sessionToken": accessToken, "refreshToken": refreshToken, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt})
	}
}

func usernameAvailability(service registration.Service, limiter *requestLimiter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		username := request.PathValue("username")
		if !usernamePattern.MatchString(username) {
			writeValidationProblem(writer)
			return
		}
		if allowed, retryAfter := limiter.allow(clientIP(request)); !allowed {
			writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeProblem(writer, http.StatusTooManyRequests, "Demasiadas consultas de username")
			return
		}

		available, err := service.UsernameAvailable(request.Context(), username)
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo consultar el username")
			return
		}
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]bool{"available": available})
	}
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return host
}

type verificationRequest struct {
	Token            string `json:"token"`
	SessionTransport string `json:"sessionTransport"`
}

func verifyRegistration(service registration.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body verificationRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || decoder.Decode(&struct{}{}) != io.EOF || (body.SessionTransport != "cookie" && body.SessionTransport != "bearer") || body.Token == "" {
			writeValidationProblem(writer)
			return
		}
		previousSession, _ := sessionToken(request)
		session, sessionToken, refreshToken, err := service.Verify(request.Context(), body.Token, previousSession.token)
		if errors.Is(err, registration.ErrVerificationInvalid) {
			writeProblem(writer, http.StatusConflict, "Verificación no válida")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo verificar la cuenta")
			return
		}
		response := map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt, "refreshExpiresAt": session.RefreshExpiresAt}
		if body.SessionTransport == "cookie" {
			http.SetCookie(writer, &http.Cookie{Name: "__Host-tm_session", Value: sessionToken, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
		} else {
			response["sessionToken"] = sessionToken
			response["refreshToken"] = refreshToken
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(response)
	}
}

func followLeague(service leagues.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		leagueID := request.PathValue("leagueId")
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo resolver la sesión")
			return
		}
		if !uuidPattern.MatchString(leagueID) {
			writeValidationProblem(writer)
			return
		}
		var err error
		if request.Method == http.MethodPut {
			err = service.Follow(request.Context(), accountID, leagueID)
		} else {
			err = service.Unfollow(request.Context(), accountID, leagueID)
		}
		if errors.Is(err, leagues.ErrLeagueNotFound) {
			writeProblem(writer, http.StatusNotFound, "Liga no disponible")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo actualizar el seguimiento")
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func listAccountLeagues(service leagues.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accountID, authenticated := currentAccountID(request.Context())
		if !authenticated {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo resolver la sesión")
			return
		}

		relationship := leagues.Relationship(request.URL.Query().Get("relationship"))
		limit, valid := listLimit(request.URL.Query().Get("limit"))
		cursor := request.URL.Query().Get("cursor")
		if !valid || (cursor != "" && !uuidPattern.MatchString(cursor)) {
			writeValidationProblem(writer)
			return
		}
		page, err := service.List(request.Context(), accountID, relationship, cursor, limit)
		if errors.Is(err, leagues.ErrInvalidRelationship) || errors.Is(err, leagues.ErrInvalidPage) {
			writeValidationProblem(writer)
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudieron consultar las ligas")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(page)
	}
}

func listLimit(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	limit, err := strconv.Atoi(raw)
	return limit, err == nil
}

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func register(service registration.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var body registerRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeValidationProblem(writer)
			return
		}

		input := registration.NormalizeInput(registration.Input{
			Email:    body.Email,
			Password: body.Password,
			Username: body.Username,
		})
		if !validRegistration(input) {
			writeValidationProblem(writer)
			return
		}
		if err := service.Register(request.Context(), input); err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo completar el registro")
			return
		}
		writer.WriteHeader(http.StatusAccepted)
	}
}

func validRegistration(input registration.Input) bool {
	if len(input.Email) == 0 || len(input.Email) > 254 || len(input.Password) < 12 || len(input.Password) > 1024 || !usernamePattern.MatchString(input.Username) {
		return false
	}
	address, err := mail.ParseAddress(input.Email)
	return err == nil && address.Address == input.Email && strings.Contains(input.Email, "@")
}

func writeValidationProblem(writer http.ResponseWriter) {
	writeProblem(writer, http.StatusBadRequest, "La solicitud no es válida")
}

func writeProblem(writer http.ResponseWriter, status int, title string) {
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"type":   "about:blank",
		"title":  title,
		"status": status,
	})
}

type requestLimiter struct {
	entries map[string]requestLimit
	limit   int
	mu      sync.Mutex
	window  time.Duration
}

type requestLimit struct {
	count   int
	resetAt time.Time
}

func newRequestLimiter(limit int, window time.Duration) *requestLimiter {
	return &requestLimiter{entries: make(map[string]requestLimit), limit: limit, window: window}
}

func (l *requestLimiter) allow(key string) (bool, int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	for entryKey, entry := range l.entries {
		if !entry.resetAt.After(now) {
			delete(l.entries, entryKey)
		}
	}

	entry := l.entries[key]
	if entry.resetAt.IsZero() {
		entry.resetAt = now.Add(l.window)
	}
	if entry.count >= l.limit {
		retryAfter := int(time.Until(entry.resetAt).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}
