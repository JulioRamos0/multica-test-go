package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JulioRamos0/multica-test-go/health"
)

func TestVersionEndpoint(t *testing.T) {
	tests := []struct {
		id         string
		method     string
		wantStatus int
		validate   func(t *testing.T, resp *http.Response, body []byte)
	}{
		{
			id:         "TC-01: Petición Exitosa (Happy Path)",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response, body []byte) {
				var r health.VersionResponse
				if err := json.Unmarshal(body, &r); err != nil {
					t.Fatalf("error al parsear JSON: %v", err)
				}
				if r.Version == "" {
					t.Errorf("se esperaba una versión no vacía, obtenido: %q", r.Version)
				}
			},
		},
		{
			id:         "TC-02: Método no permitido",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			validate:   nil,
		},
		{
			id:         "TC-03: Método PUT no permitido",
			method:     http.MethodPut,
			wantStatus: http.StatusMethodNotAllowed,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health/version", nil)
			w := httptest.NewRecorder()

			// El código productivo debe implementar la función VersionHandler
			// que maneja la ruta /health/version.
			handler := http.HandlerFunc(health.VersionHandler)
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
