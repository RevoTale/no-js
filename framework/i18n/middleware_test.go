package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddlewareRewriteAndContext(t *testing.T) {
	t.Parallel()

	resolver, err := NewResolver(Config{
		Locales:       []string{"en", "uk"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	})
	require.NoError(t, err)

	var gotPath string
	var gotLocale string
	handler := Middleware(MiddlewareConfig{
		Resolver: resolver,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotLocale = LocaleFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/uk/note/hello?x=1", nil)
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/note/hello", gotPath)
	require.Equal(t, "uk", gotLocale)
}

func TestMiddlewareCanonicalRedirectAndUnknownLocale(t *testing.T) {
	t.Parallel()

	resolver, err := NewResolver(Config{
		Locales:       []string{"en", "uk"},
		DefaultLocale: "en",
		PrefixMode:    PrefixAsNeeded,
	})
	require.NoError(t, err)

	handler := Middleware(MiddlewareConfig{
		Resolver: resolver,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recRedirect := httptest.NewRecorder()
	reqRedirect := httptest.NewRequest(http.MethodGet, "/en/note/hello?x=1", nil)
	handler.ServeHTTP(recRedirect, reqRedirect)
	require.Equal(t, http.StatusPermanentRedirect, recRedirect.Code)
	require.Equal(t, "/note/hello?x=1", recRedirect.Header().Get("Location"))

	recNotFound := httptest.NewRecorder()
	reqNotFound := httptest.NewRequest(http.MethodGet, "/it/note/hello", nil)
	handler.ServeHTTP(recNotFound, reqNotFound)
	require.Equal(t, http.StatusNotFound, recNotFound.Code)
}
