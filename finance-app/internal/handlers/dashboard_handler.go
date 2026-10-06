package handlers

import (
	"errors"
	"log"
	"net/http"

	"finance-app/internal/models"
	"finance-app/internal/services"
)

type DashboardHandler struct{ service *services.DashboardService }

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard", h.GetDashboard)
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	p := newQueryParser(r)
	from, to := p.date("from", false), p.date("to", true)
	if p.err != nil {
		writeServiceError(w, p.err, "dashboard")
		return
	}
	dashboard, err := h.service.GetDashboard(from, to)
	if err != nil {
		var validation *models.ValidationError
		if errors.As(err, &validation) {
			writeServiceError(w, err, "dashboard")
		} else {
			log.Printf("Dashboard failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}
