package handlers

import (
	"net/http"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/services"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/transactions", h.GetAll)
	mux.HandleFunc("GET /api/transactions/{id}", h.GetByID)
	mux.HandleFunc("POST /api/transactions", h.Create)
	mux.HandleFunc("PUT /api/transactions/{id}", h.Update)
	mux.HandleFunc("DELETE /api/transactions/{id}", h.Delete)
}

// Only writable fields are decoded; client IDs and timestamps are ignored.
type transactionRequest struct {
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CategoryID  int64     `json:"categoryId"`
	AccountID   int64     `json:"accountId"`
	Date        time.Time `json:"date"`
}

func decodeTransaction(w http.ResponseWriter, r *http.Request) (models.Transaction, bool) {
	input, ok := decodeJSONBody[transactionRequest](w, r)
	if !ok {
		return models.Transaction{}, false
	}
	return models.Transaction{Type: input.Type, Amount: input.Amount, Description: input.Description, CategoryID: input.CategoryID, AccountID: input.AccountID, Date: input.Date}, true
}

func (h *TransactionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	query, err := parseTransactionQuery(r)
	if err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	txs, err := h.service.Query(query)
	if err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	writeJSON(w, http.StatusOK, txs)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transaction")
	if !ok {
		return
	}
	tx, err := h.service.GetByID(id)
	if err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	tx, ok := decodeTransaction(w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(tx)
	if err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transaction")
	if !ok {
		return
	}
	tx, ok := decodeTransaction(w, r)
	if !ok {
		return
	}
	updated, err := h.service.Update(id, tx)
	if err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceID(w, r, "transaction")
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		writeServiceError(w, err, "transaction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
