package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/g-linville/budgeting/internal/database"
	"github.com/g-linville/budgeting/internal/handlers"
	"github.com/g-linville/budgeting/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize database
	db, err := database.InitDB("./budgeting.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Parse templates with custom functions
	funcMap := template.FuncMap{
		"formatCents": utils.CentsToUSD,
		"divf": func(a int, b float64) float64 {
			return float64(a) / b
		},
		"derefUint": func(p *uint) uint {
			if p == nil {
				return 0
			}
			return *p
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	templates, err = templates.ParseGlob("web/templates/partials/*.html")
	if err != nil {
		log.Fatalf("Failed to parse partial templates: %v", err)
	}

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize handlers with DB dependency and templates
	h := handlers.New(db, templates)

	// Static files
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Register routes
	r.Get("/health", h.HealthCheck)
	r.Get("/", h.Dashboard)

	// Expense routes
	r.Post("/expenses", h.CreateExpense)
	r.Get("/expenses/{id}/edit", h.GetExpenseEditForm)
	r.Put("/expenses/{id}", h.UpdateExpense)
	r.Delete("/expenses/{id}", h.DeleteExpense)

	// Income routes
	r.Post("/incomes", h.CreateIncome)
	r.Get("/incomes/{id}/edit", h.GetIncomeEditForm)
	r.Put("/incomes/{id}", h.UpdateIncome)
	r.Delete("/incomes/{id}", h.DeleteIncome)

	// Category routes
	r.Get("/categories", h.ListCategories)
	r.Post("/categories", h.CreateCategory)
	r.Get("/categories/{id}/edit", h.GetCategoryEditForm)
	r.Put("/categories/{id}", h.UpdateCategory)
	r.Delete("/categories/{id}", h.DeleteCategory)

	// Partial routes
	r.Get("/partials/recent-transactions", h.GetRecentTransactions)
	r.Get("/partials/overview", h.GetOverview)

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
