package health

import (
	"encoding/json"
	"net/http"
	"time"

	"erp/internal/platform/config"
	"github.com/go-chi/chi/v5"
)

type Status string

const (
	StatusUP   Status = "UP"
	StatusDOWN Status = "DOWN"
)

type HealthResponse struct {
	Status        Status    `json:"status"`
	Service       string    `json:"service"`
	Version       string    `json:"version"`
	Environment   string    `json:"environment"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
}

type Checker interface {
	Check() error
}

type Handler struct {
	cfg       *config.Config
	startTime time.Time
	checkers  []Checker
}

func NewHandler(cfg *config.Config, checkers ...Checker) *Handler {
	return &Handler{
		cfg:       cfg,
		startTime: time.Now(),
		checkers:  checkers,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", h.Liveness)
	r.Get("/readyz", h.Readiness)
	return r
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:        StatusUP,
		Service:       h.cfg.AppName,
		Version:       h.cfg.AppVersion,
		Environment:   string(h.cfg.Environment),
		UptimeSeconds: time.Since(h.startTime).Seconds(),
		Timestamp:     time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	status := StatusUP
	statusCode := http.StatusOK

	for _, checker := range h.checkers {
		if err := checker.Check(); err != nil {
			status = StatusDOWN
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	resp := HealthResponse{
		Status:        status,
		Service:       h.cfg.AppName,
		Version:       h.cfg.AppVersion,
		Environment:   string(h.cfg.Environment),
		UptimeSeconds: time.Since(h.startTime).Seconds(),
		Timestamp:     time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
