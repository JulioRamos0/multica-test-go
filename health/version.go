package health

import (
	"encoding/json"
	"net/http"
)

// VersionResponse define el contrato de la respuesta para el endpoint de versión.
type VersionResponse struct {
	Version string `json:"version"`
}

// Handler define la interfaz para los controladores de health check.
type Handler interface {
	GetVersion() VersionResponse
}

// VersionHandler maneja las peticiones al endpoint /health/version.
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resp := VersionResponse{
		Version: "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
