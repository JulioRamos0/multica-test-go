package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockRepository is a test implementation of Repository interface
type mockRepository struct {
	todos       map[string]Todo
	createErr   error
	updateErr   error
	deleteErr   error
	getErr      error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		todos: make(map[string]Todo),
	}
}

func (m *mockRepository) GetAll(ctx context.Context) ([]Todo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var list []Todo
	for _, t := range m.todos {
		list = append(list, t)
	}
	return list, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (Todo, bool, error) {
	if m.getErr != nil {
		return Todo{}, false, m.getErr
	}
	t, ok := m.todos[id]
	return t, ok, nil
}

func (m *mockRepository) Create(ctx context.Context, t Todo) (Todo, error) {
	if m.createErr != nil {
		return Todo{}, m.createErr
	}
	if t.ID == "" {
		t.ID = "generated-id" // mock ID generation
	}
	m.todos[t.ID] = t
	return t, nil
}

func (m *mockRepository) Update(ctx context.Context, t Todo) (Todo, bool, error) {
	if m.updateErr != nil {
		return Todo{}, false, m.updateErr
	}
	_, exists := m.todos[t.ID]
	if !exists {
		return Todo{}, false, nil
	}
	m.todos[t.ID] = t
	return t, true, nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) (bool, error) {
	if m.deleteErr != nil {
		return false, m.deleteErr
	}
	_, exists := m.todos[id]
	if !exists {
		return false, nil
	}
	delete(m.todos, id)
	return true, nil
}

// CA-01: Listado de Todos (GET)
func TestHandler_GetAll(t *testing.T) {
	repo := newMockRepository()
	repo.todos["1"] = Todo{ID: "1", Title: "Task 1", Date: time.Now()}
	repo.todos["2"] = Todo{ID: "2", Title: "Task 2", Date: time.Now()}

	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	var res []Todo
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("expected 2 items, got %v", len(res))
	}

	// Empty case
	repoEmpty := newMockRepository()
	handlerEmpty := NewHandler(repoEmpty)
	reqEmpty := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
	rrEmpty := httptest.NewRecorder()
	handlerEmpty.ServeHTTP(rrEmpty, reqEmpty)

	if rrEmpty.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rrEmpty.Code)
	}
	var resEmpty []Todo
	if err := json.NewDecoder(rrEmpty.Body).Decode(&resEmpty); err != nil {
		t.Fatalf("could not decode empty response: %v", err)
	}
	if len(resEmpty) != 0 {
		t.Errorf("expected 0 items, got %v", len(resEmpty))
	}
}

// CA-02: Consulta de Todo por ID (GET)
func TestHandler_GetByID(t *testing.T) {
	repo := newMockRepository()
	repo.todos["todo-123"] = Todo{ID: "todo-123", Title: "Specific Task", Date: time.Now()}
	handler := NewHandler(repo)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"Existing ID", "todo-123", http.StatusOK},
		{"Non-existing ID", "inexistente", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/todo?id="+tc.id, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %v, got %v", tc.wantStatus, rr.Code)
			}
		})
	}
}

// CA-03: Creación de Todo (POST)
func TestHandler_CreateTodo(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	validPayload := CreateTodoRequest{
		Title:       "Comprar insumos",
		Description: "Café",
		Date:        time.Now(),
	}

	invalidPayload := CreateTodoRequest{
		Description: "Sin título",
	}

	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
	}{
		{"Valid request", validPayload, http.StatusCreated},
		{"Missing title", invalidPayload, http.StatusBadRequest},
		{"Malformed JSON", "bad json", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			if str, ok := tc.payload.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tc.payload)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/todo", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %v, got %v", tc.wantStatus, rr.Code)
			}
		})
	}
}

// CA-04: Actualización de Todo (PUT)
func TestHandler_UpdateTodo(t *testing.T) {
	repo := newMockRepository()
	repo.todos["todo-123"] = Todo{ID: "todo-123", Title: "Old Title"}
	handler := NewHandler(repo)

	validUpdate := UpdateTodoRequest{
		ID:    "todo-123",
		Title: "New Title",
	}

	notFoundUpdate := UpdateTodoRequest{
		ID:    "no-existe",
		Title: "New Title",
	}

	tests := []struct {
		name       string
		idQuery    string
		payload    interface{}
		wantStatus int
	}{
		{"Valid update query param", "todo-123", validUpdate, http.StatusOK},
		{"Valid update body ID", "", validUpdate, http.StatusOK},
		{"Not found update", "no-existe", notFoundUpdate, http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tc.payload)
			target := "/api/todo"
			if tc.idQuery != "" {
				target += "?id=" + tc.idQuery
			}
			req := httptest.NewRequest(http.MethodPut, target, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %v, got %v", tc.wantStatus, rr.Code)
			}
		})
	}
}

// CA-05: Eliminación de Todo (DELETE)
func TestHandler_DeleteTodo(t *testing.T) {
	repo := newMockRepository()
	repo.todos["todo-123"] = Todo{ID: "todo-123", Title: "To be deleted"}
	handler := NewHandler(repo)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"Existing ID", "todo-123", http.StatusNoContent},
		{"Missing ID", "", http.StatusBadRequest},
		{"Non-existing ID", "inexistente", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := "/api/todo"
			if tc.id != "" {
				target += "?id=" + tc.id
			}
			req := httptest.NewRequest(http.MethodDelete, target, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("expected status %v, got %v", tc.wantStatus, rr.Code)
			}
		})
	}
}

// CA-06: Manejo de Métodos No Soportados
func TestHandler_MethodNotAllowed(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	methods := []string{http.MethodPatch, http.MethodHead, http.MethodOptions}

	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			req := httptest.NewRequest(m, "/api/todo", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %v, got %v for method %s", http.StatusMethodNotAllowed, rr.Code, m)
			}
		})
	}
}
