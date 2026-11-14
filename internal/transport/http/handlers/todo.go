package handlers

import (
	"Grippy/internal/application/todo"
	domain "Grippy/internal/domain/todo"
	"Grippy/internal/infrastructure/logging"

	"Grippy/internal/transport/http/dto"
	"Grippy/internal/transport/http/router"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type Router interface {
	Get(path string, h http.HandlerFunc, mw ...interface{})
	Post(path string, h http.HandlerFunc, mw ...interface{})
	Put(path string, h http.HandlerFunc, mw ...interface{})
}

type ToDoService interface {
	Save(ctx context.Context, input todo.CreateToDoInput) (*domain.ToDo, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) *domain.ToDo
	Update(ctx context.Context, input todo.UpdateToDoInput) (*domain.ToDo, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type ToDoHandler struct {
	service *todo.Service
	logger  *logging.Logger
}

func NewToDoHandler(service *todo.Service, logger *logging.Logger) *ToDoHandler {
	return &ToDoHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ToDoHandler) RegisterRoutes(r *router.Router) {
	r.Post("/add", h.Create)
	r.Get("/get", h.Get)
	r.Put("/update", h.Update)
	r.Delete("/delete", h.Delete)
}

func (h *ToDoHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateToDoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := todo.CreateToDoInput{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	t, err := h.service.Save(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, h.mapToResponse(t))
}

func (h *ToDoHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "missing todo id parameter")
		return
	}

	todoID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid todo id format")
		return
	}

	t, err := h.service.Get(r.Context(), todoID, userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, h.mapToResponse(t))
}

func (h *ToDoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "missing todo id parameter")
		return
	}

	todoID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid todo id format")
		return
	}

	var req dto.UpdateToDoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := todo.UpdateToDoInput{
		ID:          todoID,
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		IsCompleted: req.IsCompleted,
	}

	t, err := h.service.Update(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, h.mapToResponse(t))
}

func (h *ToDoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "missing todo id parameter")
		return
	}

	todoID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid todo id format")
		return
	}

	err = h.service.Delete(r.Context(), todoID, userID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "todo deleted successfully"})
}

func (h *ToDoHandler) mapToResponse(t *domain.ToDo) dto.ToDoResponse {
	return dto.ToDoResponse{
		ID:          t.ID(),
		UserID:      t.UserID(),
		Title:       t.Title(),
		Content:     t.Content(),
		IsCompleted: t.IsCompleted(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

func (h *ToDoHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrToDoNotFound):
		respondWithError(w, http.StatusNotFound, "todo not found")
	case errors.Is(err, domain.ErrEmptyTitle), errors.Is(err, domain.ErrEmptyContent):
		respondWithError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Errorf("unexpected error inside todo handler: %v", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")
	}
}
