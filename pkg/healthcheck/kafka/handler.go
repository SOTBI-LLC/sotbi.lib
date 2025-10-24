package kafka

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/log"
)

type HealthCheck struct {
	brokers []string
	log     log.Logger
	timeout *time.Duration
}

func New(brokers []string, log log.Logger, timeout *time.Duration) *HealthCheck {
	return &HealthCheck{
		brokers,
		log,
		timeout,
	}
}

type healthResponse struct {
	Status string `json:"status"`
}

func (hc HealthCheck) Handler(w http.ResponseWriter, r *http.Request) {
	// Контекст с небольшим таймаутом на проверку
	ctx, cancel := context.WithTimeout(r.Context(), *hc.timeout)
	defer cancel()

	// Пробуем подключиться к первому брокеру
	dialer := &kafka.Dialer{
		Timeout:   2 * time.Second,
		DualStack: true,
	}

	conn, err := dialer.DialContext(ctx, "tcp", hc.brokers[0])
	if err != nil {
		hc.log.Error("Kafka Healthcheck", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "kafka unreachable"})

		return
	}

	defer func(conn *kafka.Conn) {
		if err := conn.Close(); err != nil {
			hc.log.Error("Kafka Healthcheck: Close connection", "error", err)
		}
	}(conn)

	// Пробуем получить метаданные обо всех разделах (ReadPartitions без аргументов возвращает всё)
	if _, err := conn.ReadPartitions(); err != nil {
		hc.log.Error("Kafka Healthcheck: ReadPartitions", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "kafka unreachable"})

		return
	}

	// Если добрались сюда — Kafka доступна
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
