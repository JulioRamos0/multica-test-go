package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JulioRamos0/multica-test-go/auth"
)

// CA-AUTH-01 to CA-AUTH-04
func TestAPIKeyMiddleware(t *testing.T) {
	expectedKey := "secret-token-123"
	middleware := auth.NewMiddleware(expectedKey)

	// Dummy handler that simulates /api/todo
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	protectedHandler := middleware.Wrap(nextHandler)

	tests := []struct {
		name           string
		headerKey      string
		headerValue    string
		expectedStatus int
		expectedBody   string
		expectedErr    bool
	}{
		{
			name:           "CA-AUTH-01: No x-api-key header",
			headerKey:      "",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedErr:    true,
		},
		{
			name:           "CA-AUTH-02: Empty x-api-key header",
			headerKey:      "x-api-key",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedErr:    true,
		},
		{
			name:           "CA-AUTH-02: Incorrect x-api-key header",
			headerKey:      "x-api-key",
			headerValue:    "wrong-token",
			expectedStatus: http.StatusUnauthorized,
			expectedErr:    true,
		},
		{
			name:           "CA-AUTH-03: Correct x-api-key header",
			headerKey:      "x-api-key",
			headerValue:    "secret-token-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedErr:    false,
		},
		{
			name:           "CA-AUTH-04: Case-insensitive X-API-KEY header",
			headerKey:      "X-API-KEY",
			headerValue:    "secret-token-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedErr:    false,
		},
		{
			name:           "CA-AUTH-04: Case-insensitive X-Api-Key header",
			headerKey:      "X-Api-Key",
			headerValue:    "secret-token-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
			if tc.headerKey != "" {
				req.Header.Set(tc.headerKey, tc.headerValue)
			}

			rec := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.expectedErr {
				if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
					t.Errorf("expected Content-Type application/json; charset=utf-8, got %s", ct)
				}
				var resp map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse json response: %v", err)
				}
				if resp["error"] != auth.ErrMsgUnauthorized {
					t.Errorf("expected error message '%s', got '%s'", auth.ErrMsgUnauthorized, resp["error"])
				}
			} else {
				if rec.Body.String() != tc.expectedBody {
					t.Errorf("expected body '%s', got '%s'", tc.expectedBody, rec.Body.String())
				}
			}
		})
	}
}

// CA-AUTH-05
func TestMiddlewareRoutingExclusion(t *testing.T) {
	expectedKey := "secret-token-123"
	middleware := auth.NewMiddleware(expectedKey)

	mux := http.NewServeMux()
	
	// Protected route
	mux.Handle("/api/todo", middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Unprotected route (/health)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test unauthenticated to /health succeeds
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected /health to return 200 OK without header, got %d", rec.Code)
	}

	// Test unauthenticated to /api/todo fails
	reqTodo := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
	recTodo := httptest.NewRecorder()
	mux.ServeHTTP(recTodo, reqTodo)

	if recTodo.Code != http.StatusUnauthorized {
		t.Errorf("expected /api/todo to return 401 without header, got %d", recTodo.Code)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	expectedKey := "env-secret-key"
	os.Setenv(auth.EnvAppXAPIKey, expectedKey)
	defer os.Unsetenv(auth.EnvAppXAPIKey)

	config := auth.LoadConfigFromEnv()
	if config.APIKey != expectedKey {
		t.Errorf("expected config APIKey to be %s, got %s", expectedKey, config.APIKey)
	}
}
