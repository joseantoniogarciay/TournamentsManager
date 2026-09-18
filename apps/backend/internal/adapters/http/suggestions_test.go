package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/suggestions"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type suggestionRepositoryStub struct {
	item suggestions.Item
	body string
	err  error
}

func (r *suggestionRepositoryStub) CreateSuggestion(_ context.Context, _ string, body string) (suggestions.Item, error) {
	r.body = body
	return r.item, r.err
}

type suggestionNotifierStub struct{ err error }

func (n suggestionNotifierStub) NotifySuggestion(context.Context, suggestions.Item) error {
	return n.err
}

func TestSubmitSuggestionPersistsTrimmedBody(t *testing.T) {
	t.Parallel()
	repository := &suggestionRepositoryStub{item: suggestions.Item{Body: "Una mejora"}}
	request := suggestionRequestWithAccount(`{"body":"  Una mejora  "}`)
	recorder := httptest.NewRecorder()

	submitSuggestion(suggestions.NewService(repository, suggestionNotifierStub{}), newRequestLimiter(3, time.Hour)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated || repository.body != "Una mejora" {
		t.Fatalf("status = %d, body = %q; want 201 and trimmed body", recorder.Code, repository.body)
	}
}

func TestSubmitSuggestionKeepsSuccessWhenNotificationFails(t *testing.T) {
	t.Parallel()
	repository := &suggestionRepositoryStub{item: suggestions.Item{Body: "Una mejora"}}
	request := suggestionRequestWithAccount(`{"body":"Una mejora"}`)
	recorder := httptest.NewRecorder()

	submitSuggestion(suggestions.NewService(repository, suggestionNotifierStub{err: errors.New("smtp unavailable")}), newRequestLimiter(3, time.Hour)).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
}

func TestSubmitSuggestionValidatesAndRateLimitsByAccount(t *testing.T) {
	t.Parallel()
	repository := &suggestionRepositoryStub{item: suggestions.Item{Body: "Una mejora"}}
	handler := submitSuggestion(suggestions.NewService(repository, nil), newRequestLimiter(3, time.Hour))

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, suggestionRequestWithAccount(`{"body":"corto"}`))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d, want 400", invalid.Code)
	}

	for attempt := 1; attempt <= 4; attempt++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, suggestionRequestWithAccount(`{"body":"Una mejora"}`))
		want := http.StatusCreated
		if attempt == 4 {
			want = http.StatusTooManyRequests
			if recorder.Header().Get("Retry-After") == "" {
				t.Fatal("rate-limited response has no Retry-After")
			}
		}
		if recorder.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt, recorder.Code, want)
		}
	}
}

func TestSubmitSuggestionRecordsSafeRateLimitReason(t *testing.T) {
	repository := &suggestionRepositoryStub{item: suggestions.Item{Body: "Una mejora"}}
	handler := submitSuggestion(suggestions.NewService(repository, nil), newRequestLimiter(3, time.Hour))
	for attempt := 0; attempt < 3; attempt++ {
		handler.ServeHTTP(httptest.NewRecorder(), suggestionRequestWithAccount(`{"body":"Una mejora"}`))
	}

	recorder := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
	request := suggestionRequestWithAccount(`{"body":"Una mejora"}`)
	ctx, span := provider.Tracer("test").Start(request.Context(), "POST /v1/me/suggestions")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request.WithContext(ctx))
	span.End()

	spans := recorder.Ended()
	if response.Code != http.StatusTooManyRequests || len(spans) != 1 || notificationFailureReason(spans[0].Attributes()) != "rate_limit.exceeded" {
		t.Fatalf("status = %d, spans = %#v; want 429 and safe rate-limit reason", response.Code, spans)
	}
}

func TestSubmitSuggestionRecordsSafeFailureReasons(t *testing.T) {
	secretErr := errors.New("postgres failed for person@example.test with private suggestion")
	tests := []struct {
		name       string
		repository *suggestionRepositoryStub
		body       string
		wantStatus int
		wantReason string
	}{
		{"validation", &suggestionRepositoryStub{}, `{"body":"short"}`, http.StatusBadRequest, "validation.rejected"},
		{"database", &suggestionRepositoryStub{err: secretErr}, `{"body":"Una sugerencia privada"}`, http.StatusInternalServerError, "database.query_failed"},
		{"cancelled", &suggestionRepositoryStub{err: context.Canceled}, `{"body":"Una sugerencia privada"}`, http.StatusInternalServerError, "request.cancelled"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			request := suggestionRequestWithAccount(test.body)
			ctx, span := provider.Tracer("test").Start(request.Context(), "POST /v1/me/suggestions")
			response := httptest.NewRecorder()
			submitSuggestion(suggestions.NewService(test.repository, nil), newRequestLimiter(3, time.Hour)).ServeHTTP(response, request.WithContext(ctx))
			span.End()

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			spans := recorder.Ended()
			if len(spans) != 1 || notificationFailureReason(spans[0].Attributes()) != test.wantReason {
				t.Fatalf("spans = %#v, want reason %q", spans, test.wantReason)
			}
			for _, attribute := range spans[0].Attributes() {
				if strings.Contains(attribute.Value.AsString(), "person@example.test") || strings.Contains(attribute.Value.AsString(), "private suggestion") {
					t.Fatalf("attribute leaked private data: %v", attribute)
				}
			}
		})
	}
}

func suggestionRequestWithAccount(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/v1/me/suggestions", bytes.NewBufferString(body))
	return request.WithContext(context.WithValue(request.Context(), accountContextKey{}, "019abcde-1111-7111-8111-111111111111"))
}
