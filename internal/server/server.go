package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourname/metric-agent/internal/collector"
)

type Server struct {
	addr      string
	collector *collector.Collector
	srv       *http.Server
}

func New(addr string, c *collector.Collector) *Server {
	return &Server{
		addr:      addr,
		collector: c,
	}
}

func (s *Server) Start() error {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", s.healthHandler).Methods("GET")
	r.HandleFunc("/metrics", s.metricsJSONHandler).Methods("GET")
	// Prometheus default handler is registered via collector's registry; expose default via /metrics-prom
	r.Handle("/metrics-prom", http.DefaultServeMux) // prometheus handlers are mounted on DefaultServeMux

	s.srv = &http.Server{
		Addr:         s.addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		_ = s.srv.ListenAndServe()
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) metricsJSONHandler(w http.ResponseWriter, r *http.Request) {
	m := s.collector.Last()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(m)
}
