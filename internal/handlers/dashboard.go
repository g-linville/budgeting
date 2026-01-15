package handlers

import (
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/g-linville/budgeting/internal/models"
	"github.com/g-linville/budgeting/internal/utils"
)

// DashboardData holds all data needed for the dashboard template
type DashboardData struct {
	Categories         []models.Category
	RecentTransactions []Transaction
	Overview           OverviewStats
	CurrentMonth       int
	CurrentYear        int
}

// Transaction represents a combined view of expenses and income
type Transaction struct {
	ID         uint
	Type       string  // "expense" or "income"
	Name       string
	Amount     string  // Pre-formatted "$12.34"
	AmountRaw  int     // Raw cents value for calculations
	Date       string  // "2026-01-14"
	DateParsed time.Time
	Category   *string // Category name (nil for income)
	CategoryID *uint
	Notes      string
}

// OverviewStats holds summary statistics for the dashboard
type OverviewStats struct {
	TotalIncome   string // "$5,000.00"
	TotalExpenses string // "$3,245.67"
	NetSavings    string // "$1,754.33" (can be negative)
	IsPositive    bool   // Whether net savings is positive
}

// Dashboard renders the main dashboard page
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	// Query all categories for dropdown
	var categories []models.Category
	if err := h.db.Find(&categories).Error; err != nil {
		log.Printf("Error querying categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get recent transactions (last 20)
	transactions, err := h.getRecentTransactionsData(20)
	if err != nil {
		log.Printf("Error querying transactions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate overview stats for current month
	overview, err := h.calculateOverviewStats(currentMonth, currentYear)
	if err != nil {
		log.Printf("Error calculating overview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		Categories:         categories,
		RecentTransactions: transactions,
		Overview:           overview,
		CurrentMonth:       currentMonth,
		CurrentYear:        currentYear,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// getRecentTransactionsData queries and combines recent expenses and income
func (h *Handler) getRecentTransactionsData(limit int) ([]Transaction, error) {
	// Query recent expenses
	var expenses []models.Expense
	if err := h.db.Preload("Category").
		Order("expense_date DESC").
		Limit(limit).
		Find(&expenses).Error; err != nil {
		return nil, err
	}

	// Query recent income
	var incomes []models.Income
	if err := h.db.Order("income_date DESC").
		Limit(limit).
		Find(&incomes).Error; err != nil {
		return nil, err
	}

	// Convert to common Transaction type
	var transactions []Transaction

	for _, e := range expenses {
		var categoryName *string
		if e.Category != nil {
			categoryName = &e.Category.Name
		}

		transactions = append(transactions, Transaction{
			ID:         e.ID,
			Type:       "expense",
			Name:       e.Name,
			Amount:     utils.CentsToUSD(e.Amount),
			AmountRaw:  e.Amount,
			Date:       e.ExpenseDate.Format("2006-01-02"),
			DateParsed: e.ExpenseDate,
			Category:   categoryName,
			CategoryID: e.CategoryID,
			Notes:      e.Notes,
		})
	}

	for _, i := range incomes {
		transactions = append(transactions, Transaction{
			ID:         i.ID,
			Type:       "income",
			Name:       i.Name,
			Amount:     utils.CentsToUSD(i.Amount),
			AmountRaw:  i.Amount,
			Date:       i.IncomeDate.Format("2006-01-02"),
			DateParsed: i.IncomeDate,
			Category:   nil,
			CategoryID: nil,
			Notes:      i.Notes,
		})
	}

	// Sort by date (newest first)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].DateParsed.After(transactions[j].DateParsed)
	})

	// Limit to requested number
	if len(transactions) > limit {
		transactions = transactions[:limit]
	}

	return transactions, nil
}

// calculateOverviewStats calculates total income, expenses, and net savings for a given month
func (h *Handler) calculateOverviewStats(month, year int) (OverviewStats, error) {
	// Calculate date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Query expenses for the month
	var expenses []models.Expense
	if err := h.db.Where("expense_date BETWEEN ? AND ?", startDate, endDate).
		Find(&expenses).Error; err != nil {
		return OverviewStats{}, err
	}

	// Query income for the month
	var incomes []models.Income
	if err := h.db.Where("income_date BETWEEN ? AND ?", startDate, endDate).
		Find(&incomes).Error; err != nil {
		return OverviewStats{}, err
	}

	// Calculate totals
	var totalExpenses, totalIncome int
	for _, e := range expenses {
		totalExpenses += e.Amount
	}
	for _, i := range incomes {
		totalIncome += i.Amount
	}

	netSavings := totalIncome - totalExpenses

	return OverviewStats{
		TotalIncome:   utils.CentsToUSD(totalIncome),
		TotalExpenses: utils.CentsToUSD(totalExpenses),
		NetSavings:    utils.CentsToUSD(netSavings),
		IsPositive:    netSavings >= 0,
	}, nil
}
