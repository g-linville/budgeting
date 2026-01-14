package main

import (
	"log"
	"net/http"

	"github.com/g-linville/budgeting/internal/database"
	"github.com/g-linville/budgeting/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize database
	db, err := database.InitDB("./budgeting.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize handlers with DB dependency
	h := handlers.New(db)

	// Register routes
	r.Get("/health", h.HealthCheck)

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
