package handlers

import (
	"net/http"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/services"
)

type TransferHandler struct{ service *services.TransferService }

func NewTransferHandler(service *services.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

func (h *TransferHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/transfers", h.GetAll)
	mux.HandleFunc("GET /api/transfers/{id}", h.GetByID)
	mux.HandleFunc("POST /api/transfers", h.Create)
	mux.HandleFunc("PUT /api/transfers/{id}", h.Update)
	mux.HandleFunc("DELETE /api/transfers/{id}", h.Delete)
}

type transferRequest struct {
	FromAccountID int64     `json:"fromAccountId"`
	ToAccountID   int64     `json:"toAccountId"`
	Amount        int64     `json:"amount"`
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
}

func decodeTransfer(w http.ResponseWriter, r *http.Request) (models.Transfer, bool) {
	input, ok := decodeJSONBody[transferRequest](w, r)
	if !ok {
		return models.Transfer{}, false
	}
	return models.Transfer{FromAccountID: input.FromAccountID, ToAccountID: input.ToAccountID, Amount: input.Amount, Description: input.Description, Date: input.Date}, true
}

func (h *TransferHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	query, err := parseTransferQuery(r)
	if err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	transfers, err := h.service.Query(query)
	if err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func (h *TransferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transfer")
	if !ok {
		return
	}
	transfer, err := h.service.GetByID(id)
	if err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}

func (h *TransferHandler) Create(w http.ResponseWriter, r *http.Request) {
	transfer, ok := decodeTransfer(w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(transfer)
	if err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TransferHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transfer")
	if !ok {
		return
	}
	transfer, ok := decodeTransfer(w, r)
	if !ok {
		return
	}
	updated, err := h.service.Update(id, transfer)
	if err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *TransferHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transfer")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		writeServiceError(w, err, "transfer")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
