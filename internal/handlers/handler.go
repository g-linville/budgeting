package handlers

import (
	"html/template"

	"gorm.io/gorm"
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	db        *gorm.DB
	templates *template.Template
}

// New creates a new Handler with injected dependencies
func New(db *gorm.DB, templates *template.Template) *Handler {
	return &Handler{
		db:        db,
		templates: templates,
	}
}
