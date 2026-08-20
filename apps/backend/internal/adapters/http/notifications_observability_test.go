package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type notificationObservabilityRepository struct {
	listErr, countErr, markErr, deleteErr, deleteAllErr error
}

func (r notificationObservabilityRepository) ListNotifications(context.Context, string) ([]notifications.Item, error) {
	return nil, r.listErr
}

func (r notificationObservabilityRepository) UnreadCount(context.Context, string) (int, error) {
	return 0, r.countErr
}

func (r notificationObservabilityRepository) MarkAllRead(context.Context, string) error {
	return r.markErr
}

func (r notificationObservabilityRepository) Delete(context.Context, string, string) error {
	return r.deleteErr
}

func (r notificationObservabilityRepository) DeleteAll(context.Context, string) error {
	return r.deleteAllErr
}

func TestNotificationHandlersRecordOnlyClosedSafeFailures(t *testing.T) {
	secretErr := errors.New("postgres notification 019abcde-2222-7222-8222-222222222222 for person@example.test failed")
	tests := []struct {
		name       string
		handler    func(notifications.Service) http.Handler
		request    *http.Request
		repository notificationObservabilityRepository
		wantStatus int
		wantReason string
	}{
		{
			name:       "list database failure",
			handler:    func(service notifications.Service) http.Handler { return listNotifications(service) },
			request:    notificationRequest(http.MethodGet, "/v1/me/notifications"),
			repository: notificationObservabilityRepository{listErr: secretErr},
			wantStatus: http.StatusInternalServerError,
			wantReason: "database.query_failed",
		},
		{
			name:       "unread count cancelled",
			handler:    func(service notifications.Service) http.Handler { return unreadNotificationCount(service) },
			request:    notificationRequest(http.MethodGet, "/v1/me/notifications/unread-count"),
			repository: notificationObservabilityRepository{countErr: context.Canceled},
			wantStatus: http.StatusInternalServerError,
			wantReason: "request.cancelled",
		},
		{
			name:       "mark all read timeout",
			handler:    func(service notifications.Service) http.Handler { return markAllNotificationsRead(service) },
			request:    notificationRequest(http.MethodPost, "/v1/me/notifications/read"),
			repository: notificationObservabilityRepository{markErr: context.DeadlineExceeded},
			wantStatus: http.StatusInternalServerError,
			wantReason: "request.timeout",
		},
		{
			name:       "delete all database failure",
			handler:    func(service notifications.Service) http.Handler { return deleteAllNotifications(service) },
			request:    notificationRequest(http.MethodDelete, "/v1/me/notifications"),
			repository: notificationObservabilityRepository{deleteAllErr: secretErr},
			wantStatus: http.StatusInternalServerError,
			wantReason: "database.query_failed",
		},
		{
			name:       "delete one database failure",
			handler:    func(service notifications.Service) http.Handler { return deleteNotification(service) },
			request:    notificationDeleteRequest(),
			repository: notificationObservabilityRepository{deleteErr: secretErr},
			wantStatus: http.StatusInternalServerError,
			wantReason: "database.query_failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			ctx, span := provider.Tracer("test").Start(test.request.Context(), test.request.Method+" root")
			response := httptest.NewRecorder()
			test.handler(notifications.NewService(test.repository)).ServeHTTP(response, test.request.WithContext(ctx))
			span.End()

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			spans := recorder.Ended()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := notificationFailureReason(spans[0].Attributes()); got != test.wantReason {
				t.Errorf("failure reason = %q, want %q", got, test.wantReason)
			}
			for _, attribute := range spans[0].Attributes() {
				if strings.Contains(attribute.Value.AsString(), "019abcde") || strings.Contains(attribute.Value.AsString(), "person@example.test") {
					t.Fatalf("span attribute leaked request or error data: %v", attribute)
				}
			}
		})
	}
}

func TestNotificationHandlersDoNotRecordFailuresForSuccess(t *testing.T) {
	tests := []struct {
		name    string
		handler func(notifications.Service) http.Handler
		request *http.Request
	}{
		{"list", func(service notifications.Service) http.Handler { return listNotifications(service) }, notificationRequest(http.MethodGet, "/v1/me/notifications")},
		{"unread count", func(service notifications.Service) http.Handler { return unreadNotificationCount(service) }, notificationRequest(http.MethodGet, "/v1/me/notifications/unread-count")},
		{"mark all read", func(service notifications.Service) http.Handler { return markAllNotificationsRead(service) }, notificationRequest(http.MethodPost, "/v1/me/notifications/read")},
		{"delete all", func(service notifications.Service) http.Handler { return deleteAllNotifications(service) }, notificationRequest(http.MethodDelete, "/v1/me/notifications")},
		{"delete one", func(service notifications.Service) http.Handler { return deleteNotification(service) }, notificationDeleteRequest()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			ctx, span := provider.Tracer("test").Start(test.request.Context(), test.request.Method+" root")
			test.handler(notifications.NewService(notificationObservabilityRepository{})).ServeHTTP(httptest.NewRecorder(), test.request.WithContext(ctx))
			span.End()

			spans := recorder.Ended()
			if len(spans) != 1 {
				t.Fatalf("span count = %d, want 1", len(spans))
			}
			if got := notificationFailureReason(spans[0].Attributes()); got != "" {
				t.Errorf("failure reason = %q, want none", got)
			}
		})
	}
}

func notificationRequest(method, target string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	return request.WithContext(context.WithValue(request.Context(), accountContextKey{}, "019abcde-1111-7111-8111-111111111111"))
}

func notificationDeleteRequest() *http.Request {
	request := notificationRequest(http.MethodDelete, "/v1/me/notifications/019abcde-2222-7222-8222-222222222222")
	request.SetPathValue("notificationId", "019abcde-2222-7222-8222-222222222222")
	return request
}

func notificationFailureReason(attributes []attribute.KeyValue) string {
	for _, attribute := range attributes {
		if attribute.Key == "tournaments_manager.failure.reason" {
			return attribute.Value.AsString()
		}
	}
	return ""
}
