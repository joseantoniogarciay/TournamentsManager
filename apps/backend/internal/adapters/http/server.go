// Package http expone adaptadores HTTP de la API.
package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewHandler construye las rutas de infraestructura disponibles antes de los
// endpoints de negocio.
func NewHandler(registrationService registration.Service, authenticator sessionAuthenticator, leagueService leagues.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /v1/registrations", register(registrationService))
	mux.HandleFunc("POST /v1/registration-verifications", verifyRegistration(registrationService))
	mux.Handle("GET /v1/me/leagues", requireSession(authenticator)(http.HandlerFunc(listAccountLeagues(leagueService))))
	followHandler := requireSession(authenticator)(requireCookieCSRF(http.HandlerFunc(followLeague(leagueService))))
	mux.Handle("PUT /v1/me/leagues/{leagueId}/follow", followHandler)
	mux.Handle("DELETE /v1/me/leagues/{leagueId}/follow", followHandler)
	return mux
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
		session, sessionToken, err := service.Verify(request.Context(), body.Token)
		if errors.Is(err, registration.ErrVerificationInvalid) {
			writeProblem(writer, http.StatusConflict, "Verificación no válida")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "No se pudo verificar la cuenta")
			return
		}
		response := map[string]any{"user": map[string]string{"id": session.AccountID, "username": session.Username}, "delivery": body.SessionTransport, "expiresAt": session.IdleExpiresAt}
		if body.SessionTransport == "cookie" {
			http.SetCookie(writer, &http.Cookie{Name: "__Host-tm_session", Value: sessionToken, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
		} else {
			response["sessionToken"] = sessionToken
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
