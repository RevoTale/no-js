package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWithPublicFiles(t *testing.T) {
	t.Parallel()

	publicDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(publicDir, "nested"), 0o755); err != nil {
		t.Fatalf("create nested public dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(publicDir, "favicon.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatalf("write favicon fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(publicDir, "nested", "info.txt"), []byte("nested-file"), 0o644); err != nil {
		t.Fatalf("write nested fixture: %v", err)
	}

	t.Run("serves file bytes and default cache policy", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		})

		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("public file status: expected %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Body.String(); got != "<svg/>" {
			t.Fatalf("public file body: expected %q, got %q", "<svg/>", got)
		}
		if got := rec.Header().Get("Cache-Control"); got != defaultPublicFilesCachePolicy {
			t.Fatalf(
				"public file cache policy: expected %q, got %q",
				defaultPublicFilesCachePolicy,
				got,
			)
		}
		if nextCalled {
			t.Fatalf("public file request should not delegate to next handler")
		}
	})

	t.Run("serves nested path as-is", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nested/info.txt", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("nested file status: expected %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Body.String(); got != "nested-file" {
			t.Fatalf("nested file body: expected %q, got %q", "nested-file", got)
		}
	})

	t.Run("applies custom cache policy", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		cfg := PublicFilesConfig{Dir: publicDir}.WithPublicFileCachePolicy("public, max-age=600")
		middleware, err := WithPublicFiles(cfg)
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=600" {
			t.Fatalf("custom cache policy: expected %q, got %q", "public, max-age=600", got)
		}
	})

	t.Run("delegates unknown paths", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusAccepted)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing.file", nil))
		if rec.Code != http.StatusAccepted {
			t.Fatalf("unknown path should delegate: expected %d, got %d", http.StatusAccepted, rec.Code)
		}
		if !nextCalled {
			t.Fatalf("unknown path should delegate to next handler")
		}
	})

	t.Run("does not expose directory listing", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nested", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("directory path should delegate, got status %d", rec.Code)
		}
	})

	t.Run("matched file rejects non-read methods", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		if err != nil {
			t.Fatalf("build public middleware: %v", err)
		}

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/favicon.svg", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("non-read method status: expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})
}

func TestWithPublicFilesInvalidDir(t *testing.T) {
	t.Parallel()

	if _, err := WithPublicFiles(PublicFilesConfig{}); err == nil {
		t.Fatalf("expected error for empty public dir")
	}
	if _, err := WithPublicFiles(PublicFilesConfig{Dir: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Fatalf("expected error for missing public dir")
	}

	filePath := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file fixture: %v", err)
	}
	if _, err := WithPublicFiles(PublicFilesConfig{Dir: filePath}); err == nil {
		t.Fatalf("expected error for non-directory public path")
	}
}
