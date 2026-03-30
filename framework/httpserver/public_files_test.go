package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithPublicFiles(t *testing.T) {
	t.Parallel()

	publicDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(publicDir, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "favicon.svg"), []byte("<svg/>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "nested", "info.txt"), []byte("nested-file"), 0o644))

	t.Run("serves file bytes and default cache policy", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		})

		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "<svg/>", rec.Body.String())
		require.Equal(t, defaultPublicFilesCachePolicy, rec.Header().Get("Cache-Control"))
		require.False(t, nextCalled)
	})

	t.Run("serves nested path as-is", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nested/info.txt", nil))
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "nested-file", rec.Body.String())
	})

	t.Run("applies custom cache policy", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		cfg := PublicFilesConfig{Dir: publicDir}.WithPublicFileCachePolicy("public, max-age=600")
		middleware, err := WithPublicFiles(cfg)
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
		require.Equal(t, "public, max-age=600", rec.Header().Get("Cache-Control"))
	})

	t.Run("serves files under configured request path prefix only", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})
		cfg := PublicFilesConfig{Dir: publicDir}.WithRequestPathPrefix("site/")
		middleware, err := WithPublicFiles(cfg)
		require.NoError(t, err)

		handler := middleware(next)

		recPrefixed := httptest.NewRecorder()
		handler.ServeHTTP(recPrefixed, httptest.NewRequest(http.MethodGet, "/site/favicon.svg", nil))
		require.Equal(t, http.StatusOK, recPrefixed.Code)
		require.Equal(t, "<svg/>", recPrefixed.Body.String())

		recUnprefixed := httptest.NewRecorder()
		handler.ServeHTTP(recUnprefixed, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
		require.Equal(t, http.StatusAccepted, recUnprefixed.Code)
	})

	t.Run("delegates unknown paths", func(t *testing.T) {
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusAccepted)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing.file", nil))
		require.Equal(t, http.StatusAccepted, rec.Code)
		require.True(t, nextCalled)
	})

	t.Run("does not expose directory listing", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nested", nil))
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("matched file rejects non-read methods", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		middleware, err := WithPublicFiles(PublicFilesConfig{Dir: publicDir})
		require.NoError(t, err)

		handler := middleware(next)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/favicon.svg", nil))
		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}

func TestWithPublicFilesInvalidDir(t *testing.T) {
	t.Parallel()

	_, err := WithPublicFiles(PublicFilesConfig{})
	require.Error(t, err)
	_, err = WithPublicFiles(PublicFilesConfig{Dir: filepath.Join(t.TempDir(), "missing")})
	require.Error(t, err)

	filePath := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0o644))
	_, err = WithPublicFiles(PublicFilesConfig{Dir: filePath})
	require.Error(t, err)
}
