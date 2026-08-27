package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
)

const maxRISCEventBytes = 64 * 1024

type riscVerifier interface {
	Verify(context.Context, string) (federated.RISCEvent, error)
}

type riscProcessor interface {
	HandleRISCEvent(context.Context, federated.RISCEvent) error
}

// NewRISCEventReceiver builds the infrastructure webhook used exclusively by
// Google Cross-Account Protection. Google expects 202 only after a SET has been
// verified and its local session-security effect is durable.
func NewRISCEventReceiver(verifier riscVerifier, processor riscProcessor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/secevent+jwt" {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeProblem(w, http.StatusBadRequest, "Invalid security event")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRISCEventBytes))
		if err != nil || len(body) == 0 {
			observability.RecordEndpointFailure(r.Context(), "validation.rejected")
			writeProblem(w, http.StatusBadRequest, "Invalid security event")
			return
		}
		event, err := verifier.Verify(r.Context(), strings.TrimSpace(string(body)))
		if err != nil {
			if errors.Is(err, federated.ErrRISCUnavailable) {
				observability.RecordEndpointFailure(r.Context(), "identity.google_unavailable")
				writeProblem(w, http.StatusServiceUnavailable, "Security event unavailable")
				return
			}
			slog.InfoContext(r.Context(), "RISC security event rejected", "reason", riscValidationReason(err))
			observability.RecordEndpointFailure(r.Context(), "credential.risc_invalid")
			writeProblem(w, http.StatusBadRequest, "Invalid security event")
			return
		}
		if err := processor.HandleRISCEvent(r.Context(), event); err != nil {
			if errors.Is(err, federated.ErrChallengeInvalid) {
				observability.RecordEndpointFailure(r.Context(), "credential.risc_invalid")
				writeProblem(w, http.StatusBadRequest, "Invalid security event")
				return
			}
			observability.RecordEndpointFailure(r.Context(), "database.query_failed")
			writeProblem(w, http.StatusServiceUnavailable, "Security event unavailable")
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}

func riscValidationReason(err error) string {
	switch err.Error() {
	case "SET RISC mal formado":
		return "jwt.malformed"
	case "cabecera SET inválida":
		return "jwt.header_invalid"
	case "cabecera SET no permitida":
		return "jwt.header_rejected"
	case "claims SET inválidos":
		return "jwt.claims_invalid"
	case "SET RISC sin jti":
		return "jwt.jti_missing"
	case "SET RISC sin emisor":
		return "jwt.issuer_missing"
	case "SET RISC sin audiencia":
		return "jwt.audience_missing"
	case "SET RISC con número de eventos no admitido":
		return "event.count_rejected"
	case "audiencia SET no permitida":
		return "jwt.audience_rejected"
	case "emisor SET no permitido":
		return "jwt.issuer_rejected"
	case "firma SET inválida":
		return "jwt.signature_invalid"
	case "evento RISC de verificación inválido", "evento RISC inválido", "SET RISC sin evento":
		return "event.invalid"
	default:
		return "jwt.invalid"
	}
}
