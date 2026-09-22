package health

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Checker interface {
	Check() CheckResult
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type ReadinessResponse struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks,omitempty"`
}

func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:    StatusUp,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error if needed, but not required for tests
	}
}

func ReadinessHandlerFactory(checkers ...Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		overallStatus := StatusUp
		checks := make([]CheckResult, 0, len(checkers))

		for _, checker := range checkers {
			res := checker.Check()
			checks = append(checks, res)
			if res.Status == StatusDown {
				overallStatus = StatusDown
			}
		}

		response := ReadinessResponse{
			Status: overallStatus,
			Checks: checks,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if overallStatus == StatusDown {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log error if needed
		}
	}
}
