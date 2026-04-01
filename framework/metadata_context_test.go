package framework

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	frameworksite "github.com/RevoTale/no-js/framework/site"
	"github.com/stretchr/testify/require"
)

type metadataTestApp struct {
	root *url.URL
	cfg  frameworki18n.Config
}

func (app metadataTestApp) ResolveRoot(*http.Request) *url.URL {
	if app.root == nil {
		return nil
	}
	clone := *app.root
	return &clone
}

func (app metadataTestApp) I18nConfig() frameworki18n.Config {
	return app.cfg
}

type metadataResolverApp struct {
	resolver frameworksite.Resolver
	cfg      frameworki18n.Config
}

func (app metadataResolverApp) SiteResolver() frameworksite.Resolver {
	return app.resolver
}

func (app metadataResolverApp) I18nConfig() frameworki18n.Config {
	return app.cfg
}

type metadataStringResolver struct {
	canonical string
}

func (resolver metadataStringResolver) CanonicalURL() string {
	return resolver.canonical
}

func (resolver metadataStringResolver) Resolve(*http.Request) string {
	return ""
}

func TestMetadataContextProvidesRequestScopedURLHelpers(t *testing.T) {
	t.Parallel()

	meta := NewMetadataContext(
		metadataTestApp{
			root: mustParseMetadataURL(t, "https://example.com/blog"),
			cfg: frameworki18n.Config{
				Locales:       []string{"en", "de"},
				DefaultLocale: "en",
				PrefixMode:    frameworki18n.PrefixAsNeeded,
			},
		},
		newMetadataRequest(
			t,
			"https://request.example/note/hello?tag=go&__live=navigation",
			"de",
		),
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

	meta := NewMetadataContext(
		metadataTestApp{
			root: mustParseMetadataURL(t, "https://example.com/blog"),
			cfg: frameworki18n.Config{
				Locales:       []string{"en", "de"},
				DefaultLocale: "en",
				PrefixMode:    frameworki18n.PrefixAsNeeded,
			},
		},
		newMetadataRequest(
			t,
			"https://request.example/note/hello?tag=go&__live=navigation",
			"de",
		),
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

func TestMetadataContextFallsBackToSiteResolver(t *testing.T) {
	t.Parallel()

	meta := NewMetadataContext(
		metadataResolverApp{
			resolver: metadataStringResolver{canonical: "https://example.com/app"},
			cfg: frameworki18n.Config{
				Locales:       []string{"en", "de"},
				DefaultLocale: "en",
				PrefixMode:    frameworki18n.PrefixAsNeeded,
			},
		},
		newMetadataRequest(t, "https://request.example/note/hello", "en"),
	)

	require.Equal(t, "https://example.com/app", meta.Root().String())
	require.Equal(t, "https://example.com/app/note/hello", meta.CurrentURL().String())
}

func TestMetadataContextRootReturnsClone(t *testing.T) {
	t.Parallel()

	meta := NewMetadataContext(
		metadataTestApp{
			root: mustParseMetadataURL(t, "https://example.com/blog"),
			cfg: frameworki18n.Config{
				Locales:       []string{"en"},
				DefaultLocale: "en",
				PrefixMode:    frameworki18n.PrefixAsNeeded,
			},
		},
		newMetadataRequest(t, "https://request.example/", "en"),
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
