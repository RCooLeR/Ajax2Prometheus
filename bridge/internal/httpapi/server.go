package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/RCooLeR/AjaxBridge/internal/devicecatalog"
	"github.com/RCooLeR/AjaxBridge/internal/jeedom"
	"github.com/RCooLeR/AjaxBridge/internal/state"
	"github.com/RCooLeR/AjaxBridge/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type Server struct {
	addr             string
	state            *state.Engine
	store            *store.Store
	devices          *devicecatalog.Catalog
	jeedom           *jeedom.Store
	jeedomController *jeedom.Controller
	reg              *prometheus.Registry
	log              zerolog.Logger
	server           *http.Server
}

func New(addr string, stateEngine *state.Engine, store *store.Store, devices *devicecatalog.Catalog, jeedomStore *jeedom.Store, jeedomController *jeedom.Controller, reg *prometheus.Registry, log zerolog.Logger) *Server {
	return &Server{addr: addr, state: stateEngine, store: store, devices: devices, jeedom: jeedomStore, jeedomController: jeedomController, reg: reg, log: log}
}

func (s *Server) Run(ctx context.Context) error {
	router := chi.NewRouter()
	router.Get("/healthz", s.health)
	router.Get("/readyz", s.ready)
	router.Get("/state", s.currentState)
	router.Get("/events", s.events)
	router.Get("/devices", s.devicesJSON)
	if s.jeedom != nil {
		router.Get("/jeedom/devices", s.jeedomDevices)
		router.Get("/jeedom/devices/{device_slug}", s.jeedomDevice)
		router.Get("/jeedom/commands", s.jeedomCommands)
		router.Get("/jeedom/actions", s.jeedomActions)
		router.Get("/jeedom/control-audit", s.jeedomControlAudit)
		router.Post("/jeedom/devices/{device_slug}/control", s.jeedomControl)
	}
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

func (s *Server) jeedomDevices(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.jeedom.Devices())
}

func (s *Server) jeedomDevice(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "device_slug")
	device, ok := s.jeedom.Device(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(device)
}

func (s *Server) jeedomCommands(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.jeedom.Commands())
}

func (s *Server) jeedomActions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.jeedom.Actions())
}

func (s *Server) jeedomControlAudit(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsed
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.jeedom.ControlAudit(limit))
}

func (s *Server) jeedomControl(w http.ResponseWriter, r *http.Request) {
	if s.jeedomController == nil || !s.jeedomController.Enabled() {
		http.Error(w, "Jeedom controls are disabled", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Action string `json:"action"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Action == "" {
		body.Action = r.URL.Query().Get("action")
	}
	if body.Action == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}
	result, err := s.jeedomController.Execute(r.Context(), chi.URLParam(r, "device_slug"), body.Action, "http:"+r.RemoteAddr)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, jeedom.ErrControlDisabled):
			status = http.StatusServiceUnavailable
		case errors.Is(err, jeedom.ErrActionNotFound):
			status = http.StatusNotFound
		case errors.Is(err, jeedom.ErrActionDenied):
			status = http.StatusForbidden
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  err.Error(),
			"result": result,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
