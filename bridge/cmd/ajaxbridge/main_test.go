package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolveHealthcheckURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawURL   string
		httpAddr string
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:     "uses explicit url",
			rawURL:   "http://127.0.0.1:8080/healthz",
			httpAddr: ":9999",
			path:     "/readyz",
			want:     "http://127.0.0.1:8080/healthz",
		},
		{
			name:     "defaults wildcard host to loopback",
			httpAddr: ":8080",
			path:     "/readyz",
			want:     "http://127.0.0.1:8080/readyz",
		},
		{
			name:     "rewrites ipv4 wildcard host",
			httpAddr: "0.0.0.0:8080",
			path:     "healthz",
			want:     "http://127.0.0.1:8080/healthz",
		},
		{
			name:     "preserves localhost host",
			httpAddr: "localhost:8080",
			path:     "/readyz",
			want:     "http://localhost:8080/readyz",
		},
		{
			name:     "rewrites ipv6 wildcard host",
			httpAddr: "[::]:8080",
			path:     "/readyz",
			want:     "http://127.0.0.1:8080/readyz",
		},
		{
			name:    "rejects malformed explicit url",
			rawURL:  "127.0.0.1:8080/readyz",
			wantErr: true,
		},
		{
			name:     "rejects missing port",
			httpAddr: "localhost",
			path:     "/readyz",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveHealthcheckURL(tt.rawURL, tt.httpAddr, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got URL %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveHealthcheckURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveHealthcheckURL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunHealthcheck(t *testing.T) {
	t.Parallel()

	t.Run("accepts success status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		if err := runHealthcheck(context.Background(), server.URL, time.Second); err != nil {
			t.Fatalf("runHealthcheck returned error: %v", err)
		}
	})

	t.Run("rejects failure status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
		}))
		defer server.Close()

		if err := runHealthcheck(context.Background(), server.URL, time.Second); err == nil {
			t.Fatal("expected healthcheck failure, got nil")
		}
	})
}
