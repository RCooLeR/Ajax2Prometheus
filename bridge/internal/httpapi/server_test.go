package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteJSONDeclaresUTF8(t *testing.T) {
	rec := httptest.NewRecorder()

	writeJSON(rec, map[string]string{"name": "Рома"})

	if got := rec.Header().Get("Content-Type"); got != jsonContentType {
		t.Fatalf("Content-Type = %q, want %q", got, jsonContentType)
	}
	if !strings.Contains(rec.Body.String(), "Рома") {
		t.Fatalf("JSON body did not preserve UTF-8 text: %q", rec.Body.String())
	}
}

func TestLogoSkipsDirectories(t *testing.T) {
	chdir(t, t.TempDir())
	if err := os.Mkdir("logo.png", 0o755); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	(&Server{}).logo(rec, httptest.NewRequest(http.MethodGet, "/admin/logo.png", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestLogoServesFile(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "logo.png"), []byte("logo"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	(&Server{}).logo(rec, httptest.NewRequest(http.MethodGet, "/admin/logo.png", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "logo" {
		t.Fatalf("body = %q, want %q", body, "logo")
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
