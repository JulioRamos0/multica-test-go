package middleware

import (
	"encoding/json"
	"net/http"
)

// RateLimiter define el contrato para el almacenamiento y validación de límites.
type RateLimiter interface {
	// Allow evalúa si una nueva petición para un identificador dado está permitida.
	// Devuelve true si el límite no ha sido excedido; de lo contrario, devuelve false.
	Allow(identifier string) bool
}

// GlobalRateLimiter devuelve el middleware que aplica las restricciones
// usando una implementación de RateLimiter.
func GlobalRateLimiter(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := r.Header.Get("Authorization")

			if !limiter.Allow(identifier) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				response := map[string]string{
					"error":   "too_many_requests",
					"message": "Rate limit exceeded. Try again later.",
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
