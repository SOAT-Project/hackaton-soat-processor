package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsServer provides HTTP endpoints for metrics and health checks
type MetricsServer struct {
	server *http.Server
	port   int
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &MetricsServer{
		server: server,
		port:   port,
	}
}

// Start starts the metrics server
func (s *MetricsServer) Start() error {
	logger := GetLogger()
	logger.Info("starting metrics server",
		zap.Int("port", s.port),
		zap.String("metrics_endpoint", "/metrics"),
		zap.String("health_endpoint", "/health"),
		zap.String("ready_endpoint", "/ready"),
	)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops the metrics server
func (s *MetricsServer) Stop(ctx context.Context) error {
	logger := GetLogger()
	logger.Info("stopping metrics server")
	return s.server.Shutdown(ctx)
}
