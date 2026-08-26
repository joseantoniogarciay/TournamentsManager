package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/federated"
)

type riscVerifierStub struct {
	event federated.RISCEvent
	err   error
}

func (s riscVerifierStub) Verify(context.Context, string) (federated.RISCEvent, error) {
	return s.event, s.err
}

type riscProcessorStub struct {
	called bool
	err    error
}

func (s *riscProcessorStub) HandleRISCEvent(context.Context, federated.RISCEvent) error {
	s.called = true
	return s.err
}

func TestRISCEventReceiverAcceptsVerifiedEventOnlyAfterProcessing(t *testing.T) {
	processor := &riscProcessorStub{}
	handler := NewRISCEventReceiver(riscVerifierStub{event: federated.RISCEvent{ID: "event", Issuer: federated.GoogleIssuer, Subject: "subject", Type: federated.RISCSessionsRevoked}}, processor)
	request := httptest.NewRequest(http.MethodPost, "/v1/risc/events", strings.NewReader("signed-set"))
	request.Header.Set("Content-Type", "application/secevent+jwt")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusAccepted || !processor.called {
		t.Fatalf("status = %d, processed = %t", recorder.Code, processor.called)
	}
}

func TestRISCEventReceiverRejectsInvalidAndDoesNotProcess(t *testing.T) {
	processor := &riscProcessorStub{}
	handler := NewRISCEventReceiver(riscVerifierStub{err: errors.New("invalid")}, processor)
	request := httptest.NewRequest(http.MethodPost, "/v1/risc/events", strings.NewReader("bad"))
	request.Header.Set("Content-Type", "application/secevent+jwt")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || processor.called {
		t.Fatalf("status = %d, processed = %t", recorder.Code, processor.called)
	}
}

func TestRISCEventReceiverReturnsSafeUnavailableForProcessorFailure(t *testing.T) {
	processor := &riscProcessorStub{err: errors.New("database refused event")}
	handler := NewRISCEventReceiver(riscVerifierStub{event: federated.RISCEvent{ID: "event", Issuer: federated.GoogleIssuer, Subject: "subject", Type: federated.RISCSessionsRevoked}}, processor)
	request := httptest.NewRequest(http.MethodPost, "/v1/risc/events", strings.NewReader("signed-set"))
	request.Header.Set("Content-Type", "application/secevent+jwt")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable || !processor.called || strings.Contains(recorder.Body.String(), "database refused") {
		t.Fatalf("status = %d, body = %q, processed = %t", recorder.Code, recorder.Body.String(), processor.called)
	}
}

func TestRISCEventReceiverAllowsRetryWhenGoogleTrustMaterialIsUnavailable(t *testing.T) {
	processor := &riscProcessorStub{}
	handler := NewRISCEventReceiver(riscVerifierStub{err: federated.ErrRISCUnavailable}, processor)
	request := httptest.NewRequest(http.MethodPost, "/v1/risc/events", strings.NewReader("signed-set"))
	request.Header.Set("Content-Type", "application/secevent+jwt")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable || processor.called {
		t.Fatalf("status = %d, processed = %t", recorder.Code, processor.called)
	}
}
