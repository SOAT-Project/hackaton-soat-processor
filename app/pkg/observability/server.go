package observability

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsServer provides HTTP endpoints for metrics and health checks
type MetricsServer struct {
	server *http.Server
	port   int
	ready  bool
	mu     sync.RWMutex
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	ms := &MetricsServer{
		port:  port,
		ready: false,
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Simple health check endpoints (backward compatibility)
	mux.HandleFunc("/health", ms.handleHealth)
	mux.HandleFunc("/ready", ms.handleReady)

	// Kubernetes health check endpoints
	mux.HandleFunc("/processor/health/liveness", ms.handleLiveness)
	mux.HandleFunc("/processor/health/readiness", ms.handleReadiness)

	ms.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return ms
}

// handleHealth handles simple health check
func (s *MetricsServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReady handles simple readiness check
func (s *MetricsServer) handleReady(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOT READY"))
	}
}

// handleLiveness handles Kubernetes liveness probe
func (s *MetricsServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Liveness probe: checks if the application is alive
	// Always return 200 OK if the server is running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

// handleReadiness handles Kubernetes readiness probe
func (s *MetricsServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Readiness probe: checks if the application is ready to receive traffic
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready"}`))
	}
}

// SetReady marks the server as ready to receive traffic
func (s *MetricsServer) SetReady(ready bool) {
	s.mu.Lock()
	s.ready = ready
	s.mu.Unlock()

	logger := GetLogger()
	if ready {
		logger.Info("server marked as ready")
	} else {
		logger.Info("server marked as not ready")
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
		zap.String("liveness_endpoint", "/processor/health/liveness"),
		zap.String("readiness_endpoint", "/processor/health/readiness"),
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
