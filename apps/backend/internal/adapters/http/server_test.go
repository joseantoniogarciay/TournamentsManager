package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
