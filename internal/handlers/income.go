package handlers

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/g-linville/budgeting/internal/models"
	"github.com/g-linville/budgeting/internal/validation"
	"github.com/go-chi/chi/v5"
)

// CreateIncome handles POST /income
func (h *Handler) CreateIncome(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	amountStr := r.FormValue("amount")
	dateStr := r.FormValue("income_date")
	notes := r.FormValue("notes")

	// Validate input
	amountCents, date, validationErrors := validation.ValidateIncome(name, amountStr, dateStr)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
		return
	}

	// Create income record
	income := models.Income{
		Name:       name,
		Amount:     amountCents,
		IncomeDate: date,
		Notes:      notes,
	}

	if err := h.db.Create(&income).Error; err != nil {
		log.Printf("Error creating income: %v", err)
		http.Error(w, "Failed to create income", http.StatusInternalServerError)
		return
	}

	// Return updated recent transactions and overview stats (OOB)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	// Get updated data
	now := time.Now()
	transactions, err := h.getRecentTransactionsData(20)
	if err != nil {
		log.Printf("Error getting transactions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	overview, err := h.calculateOverviewStats(int(now.Month()), now.Year())
	if err != nil {
		log.Printf("Error calculating overview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		RecentTransactions: transactions,
		Overview:           overview,
		CurrentMonth:       int(now.Month()),
		CurrentYear:        now.Year(),
	}

	// Render recent transactions
	if err := h.templates.ExecuteTemplate(w, "recent-transactions", data); err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	// Render OOB overview stats
	oobBuf := new(bytes.Buffer)
	if err := h.templates.ExecuteTemplate(oobBuf, "overview-stats-oob", data); err != nil {
		log.Printf("Error executing OOB template: %v", err)
		return
	}
	w.Write(oobBuf.Bytes())
}

// GetIncomeEditForm handles GET /income/{id}/edit
func (h *Handler) GetIncomeEditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var income models.Income
	if err := h.db.First(&income, id).Error; err != nil {
		http.Error(w, "Income not found", http.StatusNotFound)
		return
	}

	data := struct {
		Income models.Income
	}{
		Income: income,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "transaction-edit-income", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UpdateIncome handles PUT /income/{id}
func (h *Handler) UpdateIncome(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	amountStr := r.FormValue("amount")
	dateStr := r.FormValue("income_date")
	notes := r.FormValue("notes")

	// Validate input
	amountCents, date, validationErrors := validation.ValidateIncome(name, amountStr, dateStr)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
		return
	}

	// Update income
	var income models.Income
	if err := h.db.First(&income, id).Error; err != nil {
		http.Error(w, "Income not found", http.StatusNotFound)
		return
	}

	income.Name = name
	income.Amount = amountCents
	income.IncomeDate = date
	income.Notes = notes

	if err := h.db.Save(&income).Error; err != nil {
		log.Printf("Error updating income: %v", err)
		http.Error(w, "Failed to update income", http.StatusInternalServerError)
		return
	}

	// Return updated transactions and overview
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	now := time.Now()
	transactions, _ := h.getRecentTransactionsData(20)
	overview, _ := h.calculateOverviewStats(int(now.Month()), now.Year())

	data := DashboardData{
		RecentTransactions: transactions,
		Overview:           overview,
		CurrentMonth:       int(now.Month()),
		CurrentYear:        now.Year(),
	}

	// Render recent transactions
	if err := h.templates.ExecuteTemplate(w, "recent-transactions", data); err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	// Render OOB overview stats
	oobBuf := new(bytes.Buffer)
	if err := h.templates.ExecuteTemplate(oobBuf, "overview-stats-oob", data); err != nil {
		log.Printf("Error executing OOB template: %v", err)
		return
	}
	w.Write(oobBuf.Bytes())
}

// DeleteIncome handles DELETE /income/{id}
func (h *Handler) DeleteIncome(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.db.Delete(&models.Income{}, id).Error; err != nil {
		log.Printf("Error deleting income: %v", err)
		http.Error(w, "Failed to delete income", http.StatusInternalServerError)
		return
	}

	// Return updated transactions and overview
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	now := time.Now()
	transactions, _ := h.getRecentTransactionsData(20)
	overview, _ := h.calculateOverviewStats(int(now.Month()), now.Year())

	data := DashboardData{
		RecentTransactions: transactions,
		Overview:           overview,
		CurrentMonth:       int(now.Month()),
		CurrentYear:        now.Year(),
	}

	// Render recent transactions
	if err := h.templates.ExecuteTemplate(w, "recent-transactions", data); err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	// Render OOB overview stats
	oobBuf := new(bytes.Buffer)
	if err := h.templates.ExecuteTemplate(oobBuf, "overview-stats-oob", data); err != nil {
		log.Printf("Error executing OOB template: %v", err)
		return
	}
	w.Write(oobBuf.Bytes())
}
