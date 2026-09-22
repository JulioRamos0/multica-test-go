package todo

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type CreateTodoRequest struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type UpdateTodoRequest struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Repository interface {
	GetAll(ctx context.Context) ([]Todo, error)
	GetByID(ctx context.Context, id string) (Todo, bool, error)
	Create(ctx context.Context, todo Todo) (Todo, error)
	Update(ctx context.Context, todo Todo) (Todo, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if id != "" {
			h.getByID(w, r, id)
			return
		}
		h.getAll(w, r)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r)
	case http.MethodDelete:
		h.delete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	todos, err := h.repo.GetAll(r.Context())
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if todos == nil {
		todos = []Todo{} // Return empty array instead of null
	}
	sendJSON(w, http.StatusOK, todos)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request, id string) {
	todo, found, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		sendError(w, http.StatusNotFound, "todo not found")
		return
	}
	sendJSON(w, http.StatusOK, todo)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		sendError(w, http.StatusBadRequest, "invalid request body: title is required")
		return
	}

	todo := Todo{
		ID:          req.ID,
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
	}

	created, err := h.repo.Create(r.Context(), todo)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, created)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		id = req.ID
	}
	if id == "" {
		sendError(w, http.StatusBadRequest, "id is required")
		return
	}

	todo := Todo{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
	}

	updated, found, err := h.repo.Update(r.Context(), todo)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		sendError(w, http.StatusNotFound, "todo not found")
		return
	}

	sendJSON(w, http.StatusOK, updated)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "id is required")
		return
	}

	found, err := h.repo.Delete(r.Context(), id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		sendError(w, http.StatusNotFound, "todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, ErrorResponse{Error: msg})
}
