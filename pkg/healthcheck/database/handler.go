package database

import (
	"context"
	"net/http"
	"time"

	"github.com/COTBU/sotbi.lib/pkg/log"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthCheck struct {
	Pinger
	log.Logger
	*time.Duration
}

func New(pinger Pinger, log log.Logger, timeout *time.Duration) *HealthCheck {
	return &HealthCheck{
		pinger,
		log,
		timeout,
	}
}

func (hc HealthCheck) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), *hc.Duration)
	defer cancel()

	err := hc.Ping(ctx)
	if err != nil {
		hc.Error("failed to ping database", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}
