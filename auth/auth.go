package auth

import (
	"encoding/json"
	"net/http"
	"os"
)

// Constantes para autenticación por API Key.
const (
	// HeaderXAPIKey nombre canónico del encabezado HTTP para la API key.
	HeaderXAPIKey = "x-api-key"
	// EnvAppXAPIKey nombre de la variable de entorno que define la clave autorizada.
	EnvAppXAPIKey = "APP_X_API_KEY"
	// ErrMsgUnauthorized mensaje devuelto en el cuerpo JSON en caso de credenciales inválidas.
	ErrMsgUnauthorized = "unauthorized"
)

// AuthConfig contiene la configuración de autenticación.
type AuthConfig struct {
	APIKey string
}

// LoadConfigFromEnv carga la clave API desde la variable de entorno APP_X_API_KEY.
func LoadConfigFromEnv() AuthConfig {
	return AuthConfig{
		APIKey: os.Getenv(EnvAppXAPIKey),
	}
}

// Middleware provee un middleware HTTP para proteger endpoints mediante x-api-key.
type Middleware struct {
	expectedKey string
}

// NewMiddleware construye una nueva instancia de Middleware con la clave esperada.
func NewMiddleware(apiKey string) *Middleware {
	return &Middleware{
		expectedKey: apiKey,
	}
}

// Wrap envuelve un http.Handler existente exigiendo autenticación válida.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(HeaderXAPIKey)
		if apiKey == "" || apiKey != m.expectedKey {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": ErrMsgUnauthorized,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// HandlerFunc envuelve un http.HandlerFunc exigiendo autenticación válida.
func (m *Middleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Wrap(next).ServeHTTP(w, r)
	}
}
