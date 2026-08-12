package http

import (
	"encoding/json"
	"net/http"

	"github.com/joseantoniogarciay/TournamentsManager/apps/backend/internal/notifications"
)

func currentAccountNotificationID(r *http.Request) (string, bool) {
	return currentAccountID(r.Context())
}

func listNotifications(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		items, err := service.List(r.Context(), accountID)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "No se pudieron consultar las notificaciones")
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
			writeProblem(w, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		count, err := service.UnreadCount(r.Context(), accountID)
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "No se pudo consultar el contador")
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
			writeProblem(w, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		if err := service.MarkAllRead(r.Context(), accountID); err != nil {
			writeProblem(w, http.StatusInternalServerError, "No se pudieron marcar las notificaciones")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func deleteAllNotifications(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		if err := service.DeleteAll(r.Context(), accountID); err != nil {
			writeProblem(w, http.StatusInternalServerError, "No se pudieron eliminar las notificaciones")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func deleteNotification(service notifications.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := currentAccountNotificationID(r)
		if !ok {
			writeProblem(w, http.StatusUnauthorized, "Sesión no válida")
			return
		}
		if err := service.Delete(r.Context(), accountID, r.PathValue("notificationId")); err != nil {
			writeProblem(w, http.StatusInternalServerError, "No se pudo eliminar la notificación")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
