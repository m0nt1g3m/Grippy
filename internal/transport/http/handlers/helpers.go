package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"Grippy/internal/transport/http/middlewares"

	"github.com/google/uuid"
)

func getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	id, ok := middlewares.GetUserID(r.Context())
	if !ok {
		return uuid.Nil, errors.New("user info not found in context")
	}
	return id, nil
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}
