package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/tournaments"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type resourceResetRepository struct {
	registration.Repository
	inspectErr error
}

func (r resourceResetRepository) InspectPasswordReset(context.Context, []byte) (string, error) {
	return "test@example.test", r.inspectErr
}
func (resourceResetRepository) ConsumePasswordReset(context.Context, []byte, string, []byte, []byte) (registration.Session, error) {
	return registration.Session{}, registration.ErrPasswordResetInvalid
}

type resourcePasswordProtector struct {
	calls   atomic.Int32
	entered chan struct{}
	release chan struct{}
}

func (p *resourcePasswordProtector) Hash(context.Context, string) (string, error) {
	p.calls.Add(1)
	if p.entered != nil {
		p.entered <- struct{}{}
		<-p.release
	}
	return "hash", nil
}
func (*resourcePasswordProtector) Verify(context.Context, string, string) bool { return false }

func resetResourceRequest(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/password-reset-confirmations", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.10:1000"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

const resetResourceBody = `{"token":"invalid","password":"correct password","sessionTransport":"bearer"}`

func TestInvalidResetTokensNeverHashAndAreRateLimited(t *testing.T) {
	p := &resourcePasswordProtector{}
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Registration: registration.NewService(resourceResetRepository{inspectErr: registration.ErrPasswordResetInvalid}, nil, p)})
	for i := 0; i < 11; i++ {
		response := resetResourceRequest(handler, resetResourceBody)
		want := 409
		if i == 10 {
			want = 429
		}
		if response.Code != want {
			t.Fatalf("request %d: %d", i, response.Code)
		}
		if want == 429 && response.Header().Get("Retry-After") == "" {
			t.Fatal("missing retry budget")
		}
	}
	if p.calls.Load() != 0 {
		t.Fatal("invalid link caused password hashing")
	}
}

func TestPasswordWorkSaturationReleasesSlots(t *testing.T) {
	p := &resourcePasswordProtector{entered: make(chan struct{}, 2), release: make(chan struct{})}
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Registration: registration.NewService(resourceResetRepository{}, nil, p)})
	done := make(chan int, 2)
	for range 2 {
		go func() { done <- resetResourceRequest(handler, resetResourceBody).Code }()
	}
	for range 2 {
		select {
		case <-p.entered:
		case <-time.After(time.Second):
			t.Fatal("hash did not start")
		}
	}
	response := resetResourceRequest(handler, resetResourceBody)
	close(p.release)
	for range 2 {
		if code := <-done; code != 409 {
			t.Fatalf("consume rejection: %d", code)
		}
	}
	if response.Code != 429 || p.calls.Load() != 2 {
		t.Fatal("saturated request reached hash")
	}
	if response = resetResourceRequest(handler, resetResourceBody); response.Code != 409 || p.calls.Load() != 3 {
		t.Fatal("slot was not released")
	}
}

func TestRequestBodyLimitsKnownAndChunkedBeforeHash(t *testing.T) {
	p := &resourcePasswordProtector{}
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Registration: registration.NewService(resourceResetRepository{}, nil, p)})
	for _, length := range []int64{-1, 2 * maxRequestBodyBytes} {
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/password-reset-confirmations", strings.NewReader(`{"token":"`+strings.Repeat("a", 2*maxRequestBodyBytes)+`","password":"correct password","sessionTransport":"bearer"}`))
		req.ContentLength = length
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != 400 || p.calls.Load() != 0 {
			t.Fatalf("excess body: %d hashes %d", response.Code, p.calls.Load())
		}
	}
}

func TestResetInspectionFailureRemainsSafeTechnicalFailure(t *testing.T) {
	p := &resourcePasswordProtector{}
	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	defer func() { _ = provider.Shutdown(context.Background()) }()
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Registration: registration.NewService(resourceResetRepository{inspectErr: errors.New("private SQL secret")}, nil, p)})
	ctx, span := provider.Tracer("test").Start(context.Background(), "HTTP")
	req := httptest.NewRequestWithContext(ctx, "POST", "/v1/password-reset-confirmations", strings.NewReader(resetResourceBody))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	span.End()
	if response.Code != 500 || strings.Contains(response.Body.String(), "private") || p.calls.Load() != 0 {
		t.Fatal("unsafe inspection failure")
	}
	if testSpanAttribute(exporter.GetSpans()[0].Attributes, "tournaments_manager.failure.reason") != "database.query_failed" {
		t.Fatal("missing technical category")
	}
}

func TestRequestBudgetNeverExtendsCallerDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	req, stop := withResourceBudget(httptest.NewRequestWithContext(ctx, "GET", "/healthz", nil))
	defer stop()
	<-req.Context().Done()
	if !errors.Is(req.Context().Err(), context.DeadlineExceeded) {
		t.Fatal("lost deadline")
	}
	req, stop = withResourceBudget(httptest.NewRequestWithContext(context.Background(), "GET", "/healthz", nil))
	defer stop()
	deadline, ok := req.Context().Deadline()
	if !ok || time.Until(deadline) > requestBudget {
		t.Fatal("missing budget")
	}
}

func TestLimiterCapacityCannotBeEvictedToBypassQuota(t *testing.T) {
	limiter := newRequestLimiter(1, time.Minute)
	for i := range 4096 {
		limiter.entries[string(rune(i))] = requestLimit{count: 1, resetAt: time.Now().Add(time.Minute)}
	}
	if allowed, _ := limiter.allow("new-key"); allowed {
		t.Fatal("unbounded limiter")
	}
	if allowed, _ := limiter.allow(string(rune(1))); allowed {
		t.Fatal("old quota evicted")
	}
}

type resourceAccountAuthenticator struct{}

func (resourceAccountAuthenticator) Authenticate(_ context.Context, token string) (string, error) {
	if token == "second-account" {
		return "019abcde-2222-7222-8222-222222222222", nil
	}
	return "019abcde-1111-7111-8111-111111111111", nil
}

func TestPersonalQuotaIsPerAccountAcrossSessions(t *testing.T) {
	service := tournaments.NewBrowsingService(browsingFake{})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: resourceAccountAuthenticator{}, TournamentBrowsing: &service})
	for i := range 62 {
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/v1/me/matches?view=pending", nil)
		token := "first-session"
		if i == 60 {
			token = "another-session"
		}
		if i == 61 {
			token = "second-account"
		}
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		want := 200
		if i == 60 {
			want = 429
		}
		if response.Code != want {
			t.Fatalf("request %d status %d want %d", i, response.Code, want)
		}
	}
}

func TestCredentialEndpointsShareAccountQuotaBeforePasswordWork(t *testing.T) {
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: resourceAccountAuthenticator{}})
	for i := range 11 {
		method, path := "POST", "/v1/me/reauthentication-tickets"
		if i%2 == 1 {
			method, path = "PUT", "/v1/me/local-credential"
		}
		request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer session")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		want := 400
		if i == 10 {
			want = 429
		}
		if response.Code != want {
			t.Fatalf("shared credential quota: %d want %d", response.Code, want)
		}
	}
}

func TestRepetitionQuotaRejectsBeforeRepositoryAndUsesSafeRootReason(t *testing.T) {
	service := tournaments.NewCreationService(repetitionHTTPRepository{err: tournaments.ErrTournamentRepetitionConflict})
	handler := NewHandlerWithConfig(HandlerConfig{}, HandlerDependencies{Authenticator: resourceAccountAuthenticator{}, TournamentCreation: &service})
	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	defer func() { _ = provider.Shutdown(t.Context()) }()
	for i := range 6 {
		ctx, span := provider.Tracer("test").Start(t.Context(), "HTTP")
		request := httptest.NewRequestWithContext(ctx, "POST", "/v1/tournaments/019abcde-2222-7222-8222-222222222222/repeat", strings.NewReader(`{"requestId":"019abcde-3333-7333-8333-333333333333","name":"Edition","teams":[{"name":"Team"}]}`))
		request.Header.Set("Authorization", "Bearer session")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		span.End()
		want := 409
		if i == 5 {
			want = 429
		}
		if response.Code != want {
			t.Fatalf("repetition quota: %d want %d", response.Code, want)
		}
	}
	spans := exporter.GetSpans()
	if testSpanAttribute(spans[5].Attributes, "tournaments_manager.failure.reason") != "rate_limit.exceeded" {
		t.Fatal("missing safe limiter reason")
	}
}
