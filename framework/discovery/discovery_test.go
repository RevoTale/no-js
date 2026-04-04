package discovery

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/RevoTale/no-js/framework"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

type testRuntime[C interface{}] struct {
	appContext C
}

func (runtime testRuntime[C]) AppContext() C {
	return runtime.appContext
}

func (testRuntime[C]) ResolveRoot(*http.Request) *url.URL {
	return nil
}

func (testRuntime[C]) I18n() *frameworki18n.Resolver {
	return nil
}

func (testRuntime[C]) IsPartialRequest(*http.Request) bool {
	return false
}

func (testRuntime[C]) RenderPage(*http.Request, http.ResponseWriter, templ.Component, metagen.Metadata) error {
	return nil
}

func (testRuntime[C]) RespondNotFound(http.ResponseWriter, *http.Request, framework.NotFoundContext) {
}

func (testRuntime[C]) RespondServerError(http.ResponseWriter, error) {}

func (testRuntime[C]) LogServerError(error) {}

func (testRuntime[C]) LogResolverTiming(framework.ResolverTiming) {}

func serveExact[C interface{}](
	runtime framework.RuntimeContext[C],
	handlers []framework.RouteHandler[C],
	w http.ResponseWriter,
	r *http.Request,
) bool {
	for _, handler := range handlers {
		if handler.TryServe(runtime, w, r) {
			return true
		}
	}
	return false
}

func TestExactHandlersRobotsRendersDocument(t *testing.T) {
	t.Parallel()

	handler := &Bundle[*struct{}]{
		Robots: func(framework.RuntimeContext[*struct{}], *http.Request) (Robots, error) {
			return Robots{
				Rules: []RobotsRule{
					{
						UserAgent: "*",
						Allow:     []string{"/"},
					},
				},
				Sitemaps: []string{"https://example.com/sitemap-index"},
			}, nil
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/robots.txt", nil)
	served := serveExact(testRuntime[*struct{}]{}, ExactHandlers(handler), rec, req)

	require.True(t, served)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, defaultRobotsCachePolicy, rec.Header().Get("Cache-Control"))
	require.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
	require.Contains(t, rec.Body.String(), "User-agent: *")
	require.Contains(t, rec.Body.String(), "Sitemap: https://example.com/sitemap-index")
}

func TestExactHandlersSitemapIndexAndChunkFallback(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	handler := &Bundle[*struct{}]{
		Sitemaps: []SitemapRoute[*struct{}]{
			{
				RoutePattern: "/",
				Sitemap: func(framework.RuntimeContext[*struct{}], *http.Request) ([]SitemapEntry, error) {
					return []SitemapEntry{{URL: "https://example.com/"}}, nil
				},
				GenerateSitemaps: func(framework.RuntimeContext[*struct{}], *http.Request) ([]SitemapID, error) {
					return []SitemapID{
						{ID: "root", Path: SitemapPath, Location: "https://example.com/sitemap.xml"},
						{ID: "note:0", Path: "/note/sitemap/0.xml", Location: "https://example.com/note/sitemap/0.xml"},
					}, nil
				},
				SitemapByID: func(
					_ framework.RuntimeContext[*struct{}],
					_ *http.Request,
					id string,
				) ([]SitemapEntry, error) {
					require.Equal(t, "note:0", id)
					return []SitemapEntry{
						{
							URL:             "https://example.com/note/hello-world",
							LastModified:    &now,
							ChangeFrequency: "weekly",
						},
					}, nil
				},
			},
		},
	}

	recIndex := httptest.NewRecorder()
	reqIndex := httptest.NewRequest(http.MethodGet, "https://example.com/sitemap-index", nil)
	servedIndex := serveExact(testRuntime[*struct{}]{}, ExactHandlers(handler), recIndex, reqIndex)
	require.True(t, servedIndex)
	require.Equal(t, http.StatusOK, recIndex.Code)
	require.Equal(t, defaultSitemapIndexCachePolicy, recIndex.Header().Get("Cache-Control"))
	require.Contains(t, recIndex.Body.String(), "<sitemapindex")
	require.Contains(t, recIndex.Body.String(), "https://example.com/sitemap.xml")
	require.Contains(t, recIndex.Body.String(), "https://example.com/note/sitemap/0.xml")

	recChunk := httptest.NewRecorder()
	reqChunk := httptest.NewRequest(http.MethodGet, "https://example.com/note/sitemap/0.xml", nil)
	servedChunk := MaybeServeSitemapChunk(testRuntime[*struct{}]{}, handler, recChunk, reqChunk)
	require.True(t, servedChunk)
	require.Equal(t, http.StatusOK, recChunk.Code)
	require.Equal(t, defaultSitemapCachePolicy, recChunk.Header().Get("Cache-Control"))
	require.Contains(t, recChunk.Body.String(), "<urlset")
	require.Contains(t, recChunk.Body.String(), "https://example.com/note/hello-world")
	require.Contains(t, recChunk.Body.String(), now.Format(time.RFC3339))
}

func TestExactHandlersSupportNestedDiscoveryRoutes(t *testing.T) {
	t.Parallel()

	handler := &Bundle[*struct{}]{
		Sitemaps: []SitemapRoute[*struct{}]{
			{
				RoutePattern: "/author/[slug]",
				Sitemap: func(framework.RuntimeContext[*struct{}], *http.Request) ([]SitemapEntry, error) {
					return []SitemapEntry{{URL: "https://example.com/author/nina"}}, nil
				},
			},
		},
		Feeds: []FeedRoute[*struct{}]{
			{
				RoutePattern: "/author/[slug]",
				Feed: func(framework.RuntimeContext[*struct{}], *http.Request) (FeedDocument, error) {
					return FeedDocument{
						Title:       "Author Feed",
						Link:        "https://example.com/author/nina",
						Description: "Latest author notes",
					}, nil
				},
			},
		},
	}

	recSitemap := httptest.NewRecorder()
	reqSitemap := httptest.NewRequest(http.MethodGet, "https://example.com/author/nina/sitemap.xml", nil)
	servedSitemap := serveExact(testRuntime[*struct{}]{}, ExactHandlers(handler), recSitemap, reqSitemap)
	require.True(t, servedSitemap)
	require.Equal(t, http.StatusOK, recSitemap.Code)
	require.Contains(t, recSitemap.Body.String(), "https://example.com/author/nina")

	recFeed := httptest.NewRecorder()
	reqFeed := httptest.NewRequest(http.MethodGet, "https://example.com/author/nina/feed.xml", nil)
	servedFeed := serveExact(testRuntime[*struct{}]{}, ExactHandlers(handler), recFeed, reqFeed)
	require.True(t, servedFeed)
	require.Equal(t, http.StatusOK, recFeed.Code)
	require.Contains(t, recFeed.Body.String(), "<title>Author Feed</title>")
}

func TestExactHandlersFeedRendersRSS(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	handler := &Bundle[*struct{}]{
		Feeds: []FeedRoute[*struct{}]{
			{
				RoutePattern: "/",
				Feed: func(framework.RuntimeContext[*struct{}], *http.Request) (FeedDocument, error) {
					return FeedDocument{
						Title:       "Example Feed",
						Link:        "https://example.com/",
						Description: "Latest posts",
						Language:    "en",
						SelfURL:     "https://example.com/feed.xml?locale=en",
						Items: []FeedItem{
							{
								Title:       "Hello",
								Link:        "https://example.com/note/hello",
								GUID:        "https://example.com/note/hello",
								Description: "Hello note",
								Author:      "Author",
								PublishedAt: &publishedAt,
								Categories:  []string{"go"},
							},
						},
					}, nil
				},
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/feed.xml?locale=en", nil)
	served := serveExact(testRuntime[*struct{}]{appContext: &struct{}{}}, ExactHandlers(handler), rec, req)

	require.True(t, served)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, defaultFeedCachePolicy, rec.Header().Get("Cache-Control"))
	require.Contains(t, rec.Header().Get("Content-Type"), "application/rss+xml")
	body := rec.Body.String()
	require.Contains(t, body, "<rss")
	require.Contains(t, body, "<title>Hello</title>")
	require.Contains(t, body, "https://example.com/feed.xml?locale=en")
	require.False(t, strings.Contains(body, "<lastBuildDate></lastBuildDate>"))
}
