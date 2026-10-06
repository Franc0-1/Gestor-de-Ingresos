package handlers

import (
	"net/http"

	"finance-app/internal/models"
	"finance-app/internal/services"
)

type CategoryHandler struct {
	service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/categories", h.GetAll)
	mux.HandleFunc("GET /api/categories/{id}", h.GetByID)
	mux.HandleFunc("POST /api/categories", h.Create)
	mux.HandleFunc("PUT /api/categories/{id}", h.Update)
	mux.HandleFunc("DELETE /api/categories/{id}", h.Delete)
}

// IDs and timestamps supplied by the client are deliberately not decoded.
type categoryRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func decodeCategory(w http.ResponseWriter, r *http.Request) (models.Category, bool) {
	input, ok := decodeJSONBody[categoryRequest](w, r)
	if !ok {
		return models.Category{}, false
	}
	return models.Category{Name: input.Name, Type: input.Type}, true
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetAll()
	if err != nil {
		writeServiceError(w, err, "category")
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "category")
	if !ok {
		return
	}
	category, err := h.service.GetByID(id)
	if err != nil {
		writeServiceError(w, err, "category")
		return
	}
	writeJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	category, ok := decodeCategory(w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(category)
	if err != nil {
		writeServiceError(w, err, "category")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "category")
	if !ok {
		return
	}
	category, ok := decodeCategory(w, r)
	if !ok {
		return
	}
	updated, err := h.service.Update(id, category)
	if err != nil {
		writeServiceError(w, err, "category")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "category")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		writeServiceError(w, err, "category")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
