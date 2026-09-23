package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockRateLimiter es una implementación de prueba para RateLimiter.
type MockRateLimiter struct {
	AllowFunc func(identifier string) bool
}

func (m *MockRateLimiter) Allow(identifier string) bool {
	if m.AllowFunc != nil {
		return m.AllowFunc(identifier)
	}
	return true
}

func TestGlobalRateLimiter(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("Permitir peticiones dentro del limite", func(t *testing.T) {
		mockLimiter := &MockRateLimiter{
			AllowFunc: func(identifier string) bool {
				if identifier != "Bearer user-token-1" {
					t.Errorf("Identificador inesperado: %s", identifier)
				}
				return true
			},
		}

		middleware := GlobalRateLimiter(mockLimiter)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer user-token-1")

		for i := 0; i < 60; i++ {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Esperaba status 200, obtuvo %d en la peticion %d", rec.Code, i+1)
			}
		}
	})

	t.Run("Rechazar la peticion 61+", func(t *testing.T) {
		reqCount := 0
		mockLimiter := &MockRateLimiter{
			AllowFunc: func(identifier string) bool {
				reqCount++
				return reqCount <= 60
			},
		}

		middleware := GlobalRateLimiter(mockLimiter)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer user-token-2")

		for i := 0; i < 60; i++ {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("Esperaba status 200, obtuvo %d", rec.Code)
			}
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("Esperaba status 429, obtuvo %d", rec.Code)
		}

		// Remover espacios extra y saltos de linea para comparar
		actualBody := strings.TrimSpace(rec.Body.String())
		
		var response map[string]string
		if err := json.Unmarshal([]byte(actualBody), &response); err != nil {
			t.Errorf("Respuesta no es un JSON valido: %s", actualBody)
		} else {
			if response["error"] != "too_many_requests" || response["message"] != "Rate limit exceeded. Try again later." {
				t.Errorf("JSON inesperado: %s", actualBody)
			}
		}
	})

	t.Run("Independencia por usuario", func(t *testing.T) {
		userCounts := make(map[string]int)
		mockLimiter := &MockRateLimiter{
			AllowFunc: func(identifier string) bool {
				userCounts[identifier]++
				return userCounts[identifier] <= 60
			},
		}

		middleware := GlobalRateLimiter(mockLimiter)
		handler := middleware(nextHandler)

		reqUserA := httptest.NewRequest(http.MethodGet, "/", nil)
		reqUserA.Header.Set("Authorization", "Bearer user-A")

		reqUserB := httptest.NewRequest(http.MethodGet, "/", nil)
		reqUserB.Header.Set("Authorization", "Bearer user-B")

		for i := 0; i < 60; i++ {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, reqUserA)
		}

		recA := httptest.NewRecorder()
		handler.ServeHTTP(recA, reqUserA)
		if recA.Code != http.StatusTooManyRequests {
			t.Errorf("Esperaba que el Usuario A recibiera 429, obtuvo %d", recA.Code)
		}

		recB := httptest.NewRecorder()
		handler.ServeHTTP(recB, reqUserB)
		if recB.Code != http.StatusOK {
			t.Errorf("Esperaba que el Usuario B recibiera 200, obtuvo %d", recB.Code)
		}
	})

	t.Run("Reinicio del limite (Window Reset)", func(t *testing.T) {
		reqCount := 0
		mockLimiter := &MockRateLimiter{
			AllowFunc: func(identifier string) bool {
				reqCount++
				// Simulamos que despues del rechazo (61), el limite se reinicia 
				// y permite la 62 (que seria la primera del nuevo minuto).
				if reqCount == 61 {
					return false
				}
				return true
			},
		}

		middleware := GlobalRateLimiter(mockLimiter)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer reset-token")

		for i := 0; i < 60; i++ {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}

		// 61 falla
		recFail := httptest.NewRecorder()
		handler.ServeHTTP(recFail, req)
		if recFail.Code != http.StatusTooManyRequests {
			t.Errorf("Esperaba 429 en la peticion 61, obtuvo %d", recFail.Code)
		}

		// 62 pasa (simulando que paso el minuto)
		recPass := httptest.NewRecorder()
		handler.ServeHTTP(recPass, req)
		if recPass.Code != http.StatusOK {
			t.Errorf("Esperaba 200 en la peticion 62, obtuvo %d", recPass.Code)
		}
	})
}
