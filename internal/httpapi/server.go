package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RCooLeR/Ajax2Prometheus/internal/devicecatalog"
	"github.com/RCooLeR/Ajax2Prometheus/internal/state"
	"github.com/RCooLeR/Ajax2Prometheus/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type Server struct {
	addr    string
	state   *state.Engine
	store   *store.Store
	devices *devicecatalog.Catalog
	reg     *prometheus.Registry
	log     zerolog.Logger
	server  *http.Server
}

func New(addr string, stateEngine *state.Engine, store *store.Store, devices *devicecatalog.Catalog, reg *prometheus.Registry, log zerolog.Logger) *Server {
	return &Server{addr: addr, state: stateEngine, store: store, devices: devices, reg: reg, log: log}
}

func (s *Server) Run(ctx context.Context) error {
	router := chi.NewRouter()
	router.Get("/healthz", s.health)
	router.Get("/readyz", s.ready)
	router.Get("/state", s.currentState)
	router.Get("/events", s.events)
	router.Get("/devices", s.devicesJSON)
	router.Handle("/metrics", promhttp.HandlerFor(s.reg, promhttp.HandlerOpts{}))

	s.server = &http.Server{
		Addr:              s.addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
	}()

	s.log.Info().Str("addr", s.addr).Msg("HTTP server started")
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) currentState(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.state.Snapshot())
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsed
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.store.ListEvents(limit))
}

func (s *Server) devicesJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.devices.Devices())
}
