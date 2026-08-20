package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/observability"
)

func currentAccountNotificationID(r *http.Request) (string, bool) {
	return currentAccountID(r.Context())
}

func listNotifications(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		items, err := service.List(r.Context(), accountID)
		recordNotificationFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve notifications")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	}
}
func unreadNotificationCount(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		count, err := service.UnreadCount(r.Context(), accountID)
		recordNotificationFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not retrieve unread count")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"count": count})
	}
}
func markAllNotificationsRead(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		err := service.MarkAllRead(r.Context(), accountID)
		recordNotificationFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not mark notifications as read")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func deleteAllNotifications(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		err := service.DeleteAll(r.Context(), accountID)
		recordNotificationFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not delete notifications")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func deleteNotification(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		err := service.Delete(r.Context(), accountID, r.PathValue("notificationId"))
		recordNotificationFailure(r.Context(), err)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "Could not delete notification")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// recordNotificationFailure keeps inbox failures on the HTTP root span without
// exporting account or notification identifiers. Inbox mutations are idempotent,
// so the module has no separate business-rejection vocabulary.
func recordNotificationFailure(ctx context.Context, err error) {
	if err != nil {
		observability.RecordDatabaseEndpointFailure(ctx, err)
	}
}
