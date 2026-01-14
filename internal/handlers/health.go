package handlers

import (
	"net/http"

	"github.com/g-linville/budgeting/internal/database"
)

// HealthCheck verifies that the server and database are operational
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check DB connectivity
	if err := database.Ping(h.db); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unhealthy"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
