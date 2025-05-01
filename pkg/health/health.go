package health

import (
	"net/http"
	"sync/atomic"
)

// HealthChecker manages server health state
type HealthChecker struct {
	// ready is an atomic flag that indicates readiness state
	ready atomic.Bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{}
	// Set ready to false initially
	hc.ready.Store(false)
	return hc
}

// SetReady sets the readiness state
func (hc *HealthChecker) SetReady(ready bool) {
	hc.ready.Store(ready)
}

// IsReady returns the current readiness state
func (hc *HealthChecker) IsReady() bool {
	return hc.ready.Load()
}

// LivenessHandler returns an HTTP handler for liveness checks
// Liveness checks only verify that the server is responding
func (hc *HealthChecker) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

// ReadinessHandler returns an HTTP handler for readiness checks
// Readiness checks verify that the server is ready to receive requests
func (hc *HealthChecker) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hc.IsReady() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("not ready"))
		}
	})
}

// AttachHealthEndpoints attaches health check endpoints to the given ServeMux
func AttachHealthEndpoints(mux *http.ServeMux, checker *HealthChecker) {
	mux.Handle("/healthz", checker.LivenessHandler())
	mux.Handle("/readyz", checker.ReadinessHandler())
}
