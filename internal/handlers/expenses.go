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

// CreateExpense handles POST /expenses
func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	amountStr := r.FormValue("amount")
	dateStr := r.FormValue("expense_date")
	notes := r.FormValue("notes")
	categoryIDStr := r.FormValue("category_id")

	// Validate input
	amountCents, date, validationErrors := validation.ValidateExpense(name, amountStr, dateStr)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("HX-Retarget", "#expense-form-errors")
		w.Header().Set("HX-Reswap", "innerHTML")
		w.WriteHeader(http.StatusBadRequest)
		h.templates.ExecuteTemplate(w, "validation-errors", validationErrors)
		return
	}

	// Parse category ID (optional)
	var categoryID *uint
	if categoryIDStr != "" {
		id, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			categoryIDUint := uint(id)
			categoryID = &categoryIDUint
		}
	}

	// Create expense record
	expense := models.Expense{
		Name:        name,
		Amount:      amountCents,
		CategoryID:  categoryID,
		ExpenseDate: date,
		Notes:       notes,
	}

	if err := h.db.Create(&expense).Error; err != nil {
		log.Printf("Error creating expense: %v", err)
		http.Error(w, "Failed to create expense", http.StatusInternalServerError)
		return
	}

	// Get updated data before writing response
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

	// Return updated recent transactions and overview stats (OOB)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

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

// GetExpenseEditForm handles GET /expenses/{id}/edit
func (h *Handler) GetExpenseEditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var expense models.Expense
	if err := h.db.Preload("Category").First(&expense, id).Error; err != nil {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	// Get categories for dropdown
	var categories []models.Category
	h.db.Find(&categories)

	data := struct {
		Expense    models.Expense
		Categories []models.Category
	}{
		Expense:    expense,
		Categories: categories,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "transaction-edit-expense", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UpdateExpense handles PUT /expenses/{id}
func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
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
	dateStr := r.FormValue("expense_date")
	notes := r.FormValue("notes")
	categoryIDStr := r.FormValue("category_id")

	// Validate input
	amountCents, date, validationErrors := validation.ValidateExpense(name, amountStr, dateStr)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
		return
	}

	// Parse category ID (optional)
	var categoryID *uint
	if categoryIDStr != "" {
		catID, err := strconv.ParseUint(categoryIDStr, 10, 32)
		if err == nil {
			categoryIDUint := uint(catID)
			categoryID = &categoryIDUint
		}
	}

	// Update expense
	var expense models.Expense
	if err := h.db.First(&expense, id).Error; err != nil {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	expense.Name = name
	expense.Amount = amountCents
	expense.CategoryID = categoryID
	expense.ExpenseDate = date
	expense.Notes = notes

	if err := h.db.Save(&expense).Error; err != nil {
		log.Printf("Error updating expense: %v", err)
		http.Error(w, "Failed to update expense", http.StatusInternalServerError)
		return
	}

	// Return updated transactions and overview
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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

// DeleteExpense handles DELETE /expenses/{id}
func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.db.Delete(&models.Expense{}, id).Error; err != nil {
		log.Printf("Error deleting expense: %v", err)
		http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
		return
	}

	// Return updated transactions and overview
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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
