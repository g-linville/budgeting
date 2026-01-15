package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

// GetRecentTransactions handles GET /partials/recent-transactions
func (h *Handler) GetRecentTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.getRecentTransactionsData(20)
	if err != nil {
		log.Printf("Error getting transactions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		RecentTransactions: transactions,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "recent-transactions", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetOverview handles GET /partials/overview
func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	// Parse month and year from query params (default to current)
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if monthStr := r.URL.Query().Get("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 1900 && y <= 2100 {
			year = y
		}
	}

	overview, err := h.calculateOverviewStats(month, year)
	if err != nil {
		log.Printf("Error calculating overview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		Overview:     overview,
		CurrentMonth: month,
		CurrentYear:  year,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "overview-stats", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
