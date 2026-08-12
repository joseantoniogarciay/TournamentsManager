package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/leagues"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/registration"
)

type testNotificationAuthenticator struct {
	testAuthenticator
	items []notifications.Item
}

func (a testNotificationAuthenticator) ListNotifications(context.Context, string) ([]notifications.Item, error) {
	return a.items, nil
}

func (testNotificationAuthenticator) UnreadCount(context.Context, string) (int, error) { return 0, nil }

func (testNotificationAuthenticator) MarkAllRead(context.Context, string) error { return nil }

func (testNotificationAuthenticator) Delete(context.Context, string, string) error { return nil }

func (testNotificationAuthenticator) DeleteAll(context.Context, string) error { return nil }

func TestListNotificationsUsesContractFieldNames(t *testing.T) {
	handler := NewHandler(
		registration.Service{},
		nil,
		testNotificationAuthenticator{
			testAuthenticator: testAuthenticator{accountID: "019abcde-1111-7111-8111-111111111111"},
			items: []notifications.Item{{
				ID:         "019abcde-2222-7222-8222-222222222222",
				Kind:       "league_administrator_assigned",
				LeagueID:   "019abcde-3333-7333-8333-333333333333",
				LeagueName: "Liga de prueba",
				CreatedAt:  "2026-08-12T20:00:00Z",
			}},
		},
		leagues.NewService(testLeagueRepository{}),
		testAllowedOrigins,
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/me/notifications", nil)
	request.Header.Set("Authorization", "Bearer session-token")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	const want = "{\"items\":[{\"id\":\"019abcde-2222-7222-8222-222222222222\",\"kind\":\"league_administrator_assigned\",\"leagueId\":\"019abcde-3333-7333-8333-333333333333\",\"leagueName\":\"Liga de prueba\",\"createdAt\":\"2026-08-12T20:00:00Z\",\"readAt\":null}]}\n"
	if body := recorder.Body.String(); body != want {
		t.Errorf("body = %s, want contract field names", body)
	}
}
