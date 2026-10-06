package handlers

import (
	"net/http"

	"finance-app/internal/models"
	"finance-app/internal/services"
)

type AccountHandler struct {
	service *services.AccountService
}

func NewAccountHandler(service *services.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/accounts", h.GetAll)
	mux.HandleFunc("GET /api/accounts/{id}", h.GetByID)
	mux.HandleFunc("POST /api/accounts", h.Create)
	mux.HandleFunc("PUT /api/accounts/{id}", h.Update)
	mux.HandleFunc("DELETE /api/accounts/{id}", h.Delete)
}

// Only writable fields are decoded. Client IDs and timestamps are ignored.
type accountRequest struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Currency       string `json:"currency"`
	InitialBalance int64  `json:"initialBalance"`
	Active         bool   `json:"active"`
}

func decodeAccount(w http.ResponseWriter, r *http.Request) (models.Account, bool) {
	input, ok := decodeJSONBody[accountRequest](w, r)
	if !ok {
		return models.Account{}, false
	}
	return models.Account{
		Name: input.Name, Type: input.Type, Currency: input.Currency,
		InitialBalance: input.InitialBalance, Active: input.Active,
	}, true
}

func accountID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return resourceID(w, r, "account")
}

func writeAccountError(w http.ResponseWriter, err error) {
	writeServiceError(w, err, "account")
}

func (h *AccountHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.service.GetAll()
	if err != nil {
		writeAccountError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	account, err := h.service.GetByID(id)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	account, ok := decodeAccount(w, r)
	if !ok {
		return
	}
	created, err := h.service.Create(account)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	account, ok := decodeAccount(w, r)
	if !ok {
		return
	}
	updated, err := h.service.Update(id, account)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
