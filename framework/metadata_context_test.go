package framework

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/stretchr/testify/require"
)

func TestMetadataContextProvidesRequestScopedURLHelpers(t *testing.T) {
	t.Parallel()

	meta := NewMetaContext(
		requestContext(nil),
		struct{}{},
		newMetadataRequest(
			t,
			"https://request.example/note/hello?tag=go&__live=navigation",
			"de",
		),
		mustParseMetadataURL(t, "https://example.com/blog"),
		mustNewMetadataResolver(t, frameworki18n.Config{
			Locales:       []string{"en", "de"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		}),
	)

	require.Equal(t, "de", meta.Locale())
	require.Equal(t, "https://example.com/blog", meta.Root().String())
	require.Equal(
		t,
		"https://example.com/blog/note/hello?__live=navigation&tag=go",
		meta.CurrentURL().String(),
	)
	require.Equal(t, "https://example.com/blog/feed.xml", meta.URL("/feed.xml").String())
	require.Equal(
		t,
		"https://example.com/blog/de/note/hello",
		meta.LocalizedURL("de", "/note/hello").String(),
	)
}

func TestMetadataContextAlternatesUseResolvedRootAndStripInternalQueryMarkers(t *testing.T) {
	t.Parallel()

	meta := NewMetaContext(
		requestContext(nil),
		struct{}{},
		newMetadataRequest(
			t,
			"https://request.example/note/hello?tag=go&__live=navigation",
			"de",
		),
		mustParseMetadataURL(t, "https://example.com/blog"),
		mustNewMetadataResolver(t, frameworki18n.Config{
			Locales:       []string{"en", "de"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		}),
	)

	alternates, err := meta.Alternates("de", map[string]string{
		"application/rss+xml": "/feed.xml?__live=navigation",
	})
	require.NoError(t, err)
	require.Equal(t, "https://example.com/blog/de/note/hello?tag=go", alternates.Canonical)
	require.Equal(t, "https://example.com/blog/note/hello?tag=go", alternates.Languages["en"])
	require.Equal(t, "https://example.com/blog/de/note/hello?tag=go", alternates.Languages["de"])
	require.Equal(t, "https://example.com/blog/feed.xml", alternates.Types["application/rss+xml"])
}

func TestMetadataContextAllowsNilRoot(t *testing.T) {
	t.Parallel()

	meta := NewMetaContext(
		requestContext(nil),
		struct{}{},
		newMetadataRequest(t, "https://request.example/note/hello", "en"),
		nil,
		mustNewMetadataResolver(t, frameworki18n.Config{
			Locales:       []string{"en", "de"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		}),
	)

	require.Nil(t, meta.Root())
	require.Nil(t, meta.CurrentURL())
}

func TestMetadataContextRootReturnsClone(t *testing.T) {
	t.Parallel()

	meta := NewMetaContext(
		requestContext(nil),
		struct{}{},
		newMetadataRequest(t, "https://request.example/", "en"),
		mustParseMetadataURL(t, "https://example.com/blog"),
		mustNewMetadataResolver(t, frameworki18n.Config{
			Locales:       []string{"en"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		}),
	)

	root := meta.Root()
	require.NotNil(t, root)
	root.Path = "/changed"

	require.Equal(t, "https://example.com/blog", meta.Root().String())
}

func mustParseMetadataURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	require.NoError(t, err)
	return parsed
}

func mustNewMetadataResolver(t *testing.T, cfg frameworki18n.Config) *frameworki18n.Resolver {
	t.Helper()

	resolver, err := frameworki18n.NewResolver(cfg)
	require.NoError(t, err)
	return resolver
}

func newMetadataRequest(t *testing.T, rawURL string, locale string) *http.Request {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, rawURL, nil)
	request = request.WithContext(
		frameworki18n.WithRequestInfo(request.Context(), frameworki18n.RequestInfo{
			Locale: locale,
		}),
	)
	return request
}
