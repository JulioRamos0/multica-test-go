package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JulioRamos0/multica-test-go/health"
)

// mockChecker implementa la interfaz health.Checker necesaria para inyectar fallos controlados.
type mockChecker struct {
	result health.CheckResult
}

func (m mockChecker) Check() health.CheckResult {
	return m.result
}

func TestHealthzEndpoint(t *testing.T) {
	tests := []struct {
		id         string
		method     string
		wantStatus int
		validate   func(t *testing.T, resp *http.Response, body []byte)
	}{
		{
			id:         "TC-01",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response, body []byte) {
				if ct := resp.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
					t.Errorf("esperado Content-Type application/json; charset=utf-8, obtenido %q", ct)
				}
				var r health.HealthResponse
				if err := json.Unmarshal(body, &r); err != nil {
					t.Fatalf("error al parsear JSON: %v", err)
				}
				if r.Status != health.StatusUp {
					t.Errorf("estado esperado UP, obtenido %q", r.Status)
				}
				if r.Timestamp.IsZero() {
					t.Errorf("expected non-zero timestamp formato RFC3339")
				}
			},
		},
		{
			id:         "TC-02",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			// El código productivo debe proveer la firma LivenessHandler que cumpla `http.HandlerFunc` o equivalente.
			handler := http.HandlerFunc(health.LivenessHandler)
			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("estado esperado %d, obtenido %d", tt.wantStatus, resp.StatusCode)
			}
			if tt.validate != nil {
				tt.validate(t, resp, w.Body.Bytes())
			}
		})
	}
}

func TestReadyEndpoint(t *testing.T) {
	tests := []struct {
		id         string
		method     string
		checkers   []health.Checker
		wantStatus int
		validate   func(t *testing.T, resp *http.Response, body []byte)
	}{
		{
			id:     "TC-03",
			method: http.MethodGet,
			checkers: []health.Checker{
				mockChecker{result: health.CheckResult{Name: "system", Status: health.StatusUp}},
			},
			wantStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response, body []byte) {
				var r health.ReadinessResponse
				if err := json.Unmarshal(body, &r); err != nil {
					t.Fatalf("error al parsear JSON: %v", err)
				}
				if r.Status != health.StatusUp {
					t.Errorf("estado esperado UP, obtenido %q", r.Status)
				}
				if len(r.Checks) < 1 {
					t.Errorf("longitud de checks esperada >= 1, obtenido %d", len(r.Checks))
				}
			},
		},
		{
			id:     "TC-04",
			method: http.MethodGet,
			checkers: []health.Checker{
				mockChecker{result: health.CheckResult{Name: "system", Status: health.StatusUp}},
				mockChecker{result: health.CheckResult{Name: "db", Status: health.StatusDown, Error: "db timeout"}},
			},
			wantStatus: http.StatusServiceUnavailable,
			validate: func(t *testing.T, resp *http.Response, body []byte) {
				var r health.ReadinessResponse
				if err := json.Unmarshal(body, &r); err != nil {
					t.Fatalf("error al parsear JSON: %v", err)
				}
				if r.Status != health.StatusDown {
					t.Errorf("estado esperado DOWN, obtenido %q", r.Status)
				}
				foundErr := false
				for _, c := range r.Checks {
					if c.Name == "db" && c.Error != "" {
						foundErr = true
						break
					}
				}
				if !foundErr {
					t.Errorf("esperado encontrar error en verificaciones (checks)")
				}
			},
		},
		{
			id:         "TC-05",
			method:     http.MethodDelete,
			checkers:   nil,
			wantStatus: http.StatusMethodNotAllowed,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/ready", nil)
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			// El código productivo debe proveer la fábrica ReadinessHandlerFactory recibiendo 0 o n checkers.
			handler := http.HandlerFunc(health.ReadinessHandlerFactory(tt.checkers...))
			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("estado esperado %d, obtenido %d", tt.wantStatus, resp.StatusCode)
			}
			if tt.validate != nil {
				tt.validate(t, resp, w.Body.Bytes())
			}
		})
	}
}
