package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

func decodeJSONBody[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var input *T
	if err := decoder.Decode(&input); err != nil || input == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return nil, false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body must contain one JSON object"})
		return nil, false
	}
	return input, true
}

func resourceID(w http.ResponseWriter, r *http.Request, resource string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid " + resource + " ID"})
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func writeServiceError(w http.ResponseWriter, err error, resource string) {
	var validation *models.ValidationError
	switch {
	case errors.As(err, &validation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": validation.Error()})
	case errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": resource + " not found"})
	default:
		log.Printf("%s operation failed: %v", resource, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
