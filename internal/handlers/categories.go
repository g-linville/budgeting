package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/g-linville/budgeting/internal/models"
	"github.com/g-linville/budgeting/internal/validation"
	"github.com/go-chi/chi/v5"
)

// ListCategories handles GET /categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	if err := h.db.Find(&categories).Error; err != nil {
		log.Printf("Error querying categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []models.Category
	}{
		Categories: categories,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "category-modal", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// CreateCategory handles POST /categories
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	color := r.FormValue("color")

	// Validate input
	validationErrors := validation.ValidateCategory(name, color)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
		return
	}

	// Check if category name already exists
	var existingCategory models.Category
	if err := h.db.Where("LOWER(name) = LOWER(?)", name).First(&existingCategory).Error; err == nil {
		http.Error(w, "Category with this name already exists", http.StatusBadRequest)
		return
	}

	// Create category
	category := models.Category{
		Name:  name,
		Color: color,
	}

	if err := h.db.Create(&category).Error; err != nil {
		log.Printf("Error creating category: %v", err)
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	// Return updated category list
	var categories []models.Category
	h.db.Find(&categories)

	data := struct {
		Categories []models.Category
	}{
		Categories: categories,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err := h.templates.ExecuteTemplate(w, "category-list", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// GetCategoryEditForm handles GET /categories/{id}/edit
func (h *Handler) GetCategoryEditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := struct {
		Category models.Category
	}{
		Category: category,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "category-edit", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UpdateCategory handles PUT /categories/{id}
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
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
	color := r.FormValue("color")

	// Validate input
	validationErrors := validation.ValidateCategory(name, color)
	if validationErrors.HasErrors() {
		log.Printf("Validation errors: %v", validationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
		return
	}

	// Check if category name already exists (excluding current category)
	var existingCategory models.Category
	if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, id).First(&existingCategory).Error; err == nil {
		http.Error(w, "Category with this name already exists", http.StatusBadRequest)
		return
	}

	// Update category
	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	category.Name = name
	category.Color = color

	if err := h.db.Save(&category).Error; err != nil {
		log.Printf("Error updating category: %v", err)
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	// Return updated category list
	var categories []models.Category
	h.db.Find(&categories)

	data := struct {
		Categories []models.Category
	}{
		Categories: categories,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "category-list", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// DeleteCategory handles DELETE /categories/{id}
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete category (expenses will have category_id set to NULL via ON DELETE SET NULL)
	if err := h.db.Delete(&models.Category{}, id).Error; err != nil {
		log.Printf("Error deleting category: %v", err)
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	// Return updated category list
	var categories []models.Category
	h.db.Find(&categories)

	data := struct {
		Categories []models.Category
	}{
		Categories: categories,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "category-list", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
