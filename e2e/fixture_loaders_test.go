package e2e

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type routePageCSSFixture struct {
	Home          responseSnapshot
	Partial       responseSnapshot
	NotFound      responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type templCSSFixture struct {
	Home          responseSnapshot
	Partial       responseSnapshot
	NotFound      responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type namespacedFixture struct {
	Dashboard     responseSnapshot
	Partial       responseSnapshot
	NotFound      responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type docsFeatureFixture struct {
	AuthorDE                  responseSnapshot
	AuthorEN                  responseSnapshot
	AuthorENRedirect          responseSnapshot
	AuthorMissing             responseSnapshot
	AuthorError               responseSnapshot
	Dashboard                 responseSnapshot
	DashboardPartial          responseSnapshot
	PingDE                    responseSnapshot
	Robots                    responseSnapshot
	Feed                      responseSnapshot
	AuthorFeed                responseSnapshot
	Sitemap                   responseSnapshot
	SitemapIndex              responseSnapshot
	SitemapChunk              responseSnapshot
	Favicon                   responseSnapshot
	SiteCSS                   responseSnapshot
	TemplCSS                  responseSnapshot
	Health                    responseSnapshot
	SiteCSSURL                string
	TemplCSSURL               string
	ExpensiveLoadsAfterFirst  int
	ExpensiveLoadsAfterSecond int
}

type groupedNamespaceFixture struct {
	Notes         responseSnapshot
	NotesPartial  responseSnapshot
	Guides        responseSnapshot
	Tags          responseSnapshot
	TagsPartial   responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type catchAllFixture struct {
	Docs          responseSnapshot
	Nested        responseSnapshot
	Partial       responseSnapshot
	Missing       responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type optionalCatchAllFixture struct {
	Root          responseSnapshot
	Nested        responseSnapshot
	Partial       responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type methodMatrixFixture struct {
	Home          responseSnapshot
	Get           responseSnapshot
	Head          responseSnapshot
	Options       responseSnapshot
	Post          responseSnapshot
	Patch         responseSnapshot
	Delete        responseSnapshot
	Put           responseSnapshot
	Missing       responseSnapshot
	Stylesheet    responseSnapshot
	StylesheetURL string
}

type prefixAlwaysFixture struct {
	RootRedirect       responseSnapshot
	HomeEN             responseSnapshot
	HomeDE             responseSnapshot
	NotFoundEN         responseSnapshot
	NotFoundDE         responseSnapshot
	PageLoadNotFoundEN responseSnapshot
	PageLoadNotFoundDE responseSnapshot
	HelpNotFoundEN     responseSnapshot
	HelpNotFoundDE     responseSnapshot
	GreetRedirect      responseSnapshot
	GreetEN            responseSnapshot
	GreetDE            responseSnapshot
	GreetPartial       responseSnapshot
	Stylesheet         responseSnapshot
	StylesheetURL      string
}

type customRuntimeFixture struct {
	Home          responseSnapshot
	Extra         responseSnapshot
	Public        responseSnapshot
	Health        responseSnapshot
	DefaultHealth responseSnapshot
	SiteCSS       responseSnapshot
	TemplCSS      responseSnapshot
	SiteCSSURL    string
	TemplCSSURL   string
}

type streamSnapshot struct {
	Status      int
	ContentType string
	FirstChunk  string
	Body        string
}

type templRulesFixture struct {
	BaseURL          string
	Card             responseSnapshot
	Panel            responseSnapshot
	Board            responseSnapshot
	Deps             responseSnapshot
	Hooks            responseSnapshot
	Vars             responseSnapshot
	Fallback         responseSnapshot
	Metadata         responseSnapshot
	MetadataPartial  responseSnapshot
	Dashboard        responseSnapshot
	DashboardPartial responseSnapshot
	TemplCSS         responseSnapshot
	MeterScript      responseSnapshot
	Stream           streamSnapshot
	StylesheetURL    string
}

var stylesheetPattern = regexp.MustCompile(`href="([^"]+/styles/templ\.css)"`)
var siteCSSPattern = regexp.MustCompile(`id="site-css"[^>]*href="([^"]+/site\.css)"`)
var expensiveCountPattern = regexp.MustCompile(`<div id="expensive-count">([^<]+)</div>`)

func loadRoutePageCSSFixture(t *testing.T) (string, routePageCSSFixture) {
	t.Helper()

	appDir, server := startPreparedFixtureWithNoJSGen(t, "routepagecssapp")

	home := requestFixture(t, server, http.MethodGet, "/", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/", nil, hxRequestOptions())
	notFound := requestFixture(t, server, http.MethodGet, "/missing", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, home.Body)
	stylesheet := requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{})

	return appDir, routePageCSSFixture{
		Home:          home,
		Partial:       partial,
		NotFound:      notFound,
		Stylesheet:    stylesheet,
		StylesheetURL: stylesheetURL,
	}
}

func loadTemplCSSFixture(t *testing.T) templCSSFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "templcssapp")

	home := requestFixture(t, server, http.MethodGet, "/", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/", nil, hxRequestOptions())
	notFound := requestFixture(t, server, http.MethodGet, "/missing", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, home.Body)
	stylesheet := requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{})

	return templCSSFixture{
		Home:          home,
		Partial:       partial,
		NotFound:      notFound,
		Stylesheet:    stylesheet,
		StylesheetURL: stylesheetURL,
	}
}

func loadNamespacedFixture(t *testing.T) namespacedFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "namespacedtemplcssapp")

	dashboard := requestFixture(t, server, http.MethodGet, "/dashboard", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/dashboard", nil, hxRequestOptions())
	notFound := requestFixture(t, server, http.MethodGet, "/unknown", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, dashboard.Body)
	stylesheet := requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{})

	return namespacedFixture{
		Dashboard:     dashboard,
		Partial:       partial,
		NotFound:      notFound,
		Stylesheet:    stylesheet,
		StylesheetURL: stylesheetURL,
	}
}

func loadDocsFeatureFixture(t *testing.T) docsFeatureFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "docsfeatureapp")
	opts := secureHostOptions("docs.example.test")

	authorDE := requestFixture(t, server, http.MethodGet, "/de/author/ada", nil, opts)
	authorENRedirect := requestFixture(t, server, http.MethodGet, "/en/author/ada", nil, opts)
	authorEN := requestFixture(t, server, http.MethodGet, "/author/ada", nil, opts)
	authorMissing := requestFixture(t, server, http.MethodGet, "/author/missing", nil, opts)
	authorError := requestFixture(t, server, http.MethodGet, "/author/boom", nil, opts)
	dashboard := requestFixture(t, server, http.MethodGet, "/de/dashboard", nil, opts)
	dashboardPartial := requestFixture(
		t,
		server,
		http.MethodGet,
		"/de/dashboard",
		nil,
		mergeOptions(opts, hxRequestOptions()),
	)
	pingDE := requestFixture(t, server, http.MethodGet, "/de/api/ping", nil, opts)
	robots := requestFixture(t, server, http.MethodGet, "/robots.txt", nil, opts)
	feed := requestFixture(t, server, http.MethodGet, "/feed.xml", nil, opts)
	authorFeed := requestFixture(t, server, http.MethodGet, "/author/ada/feed.xml", nil, opts)
	sitemap := requestFixture(t, server, http.MethodGet, "/sitemap.xml", nil, opts)
	sitemapIndex := requestFixture(t, server, http.MethodGet, "/sitemap-index.xml", nil, opts)
	sitemapChunk := requestFixture(t, server, http.MethodGet, "/sitemap/authors.xml", nil, opts)
	favicon := requestFixture(t, server, http.MethodGet, "/favicon.ico", nil, opts)
	health := requestFixture(t, server, http.MethodGet, "/healthz", nil, requestOptions{})

	siteCSSURL := extractSiteCSSURL(t, authorDE.Body)
	templCSSURL := extractStylesheetURL(t, authorDE.Body)

	return docsFeatureFixture{
		AuthorDE:                  authorDE,
		AuthorEN:                  authorEN,
		AuthorENRedirect:          authorENRedirect,
		AuthorMissing:             authorMissing,
		AuthorError:               authorError,
		Dashboard:                 dashboard,
		DashboardPartial:          dashboardPartial,
		PingDE:                    pingDE,
		Robots:                    robots,
		Feed:                      feed,
		AuthorFeed:                authorFeed,
		Sitemap:                   sitemap,
		SitemapIndex:              sitemapIndex,
		SitemapChunk:              sitemapChunk,
		Favicon:                   favicon,
		SiteCSS:                   requestFixture(t, server, http.MethodGet, siteCSSURL, nil, opts),
		TemplCSS:                  requestFixture(t, server, http.MethodGet, templCSSURL, nil, opts),
		Health:                    health,
		SiteCSSURL:                siteCSSURL,
		TemplCSSURL:               templCSSURL,
		ExpensiveLoadsAfterFirst:  extractExpensiveCount(t, authorDE.Body),
		ExpensiveLoadsAfterSecond: extractExpensiveCount(t, authorEN.Body),
	}
}

func loadGroupedNamespaceFixture(t *testing.T) groupedNamespaceFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "groupednamespaceapp")

	notes := requestFixture(t, server, http.MethodGet, "/discover/notes", nil, requestOptions{})
	notesPartial := requestFixture(t, server, http.MethodGet, "/discover/notes", nil, hxRequestOptions())
	guides := requestFixture(t, server, http.MethodGet, "/discover/guides", nil, requestOptions{})
	tags := requestFixture(t, server, http.MethodGet, "/discover/tags", nil, requestOptions{})
	tagsPartial := requestFixture(t, server, http.MethodGet, "/discover/tags", nil, hxRequestOptions())
	stylesheetURL := extractStylesheetURL(t, notes.Body)

	return groupedNamespaceFixture{
		Notes:         notes,
		NotesPartial:  notesPartial,
		Guides:        guides,
		Tags:          tags,
		TagsPartial:   tagsPartial,
		Stylesheet:    requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{}),
		StylesheetURL: stylesheetURL,
	}
}

func loadCatchAllFixture(t *testing.T) catchAllFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "catchallapp")

	docs := requestFixture(t, server, http.MethodGet, "/docs/a/b", nil, requestOptions{})
	nested := requestFixture(t, server, http.MethodGet, "/docs/alpha/beta/gamma", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/docs/a/b", nil, hxRequestOptions())
	missing := requestFixture(t, server, http.MethodGet, "/docs", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, docs.Body)

	return catchAllFixture{
		Docs:          docs,
		Nested:        nested,
		Partial:       partial,
		Missing:       missing,
		Stylesheet:    requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{}),
		StylesheetURL: stylesheetURL,
	}
}

func loadOptionalCatchAllFixture(t *testing.T) optionalCatchAllFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "optionalcatchallapp")

	root := requestFixture(t, server, http.MethodGet, "/library", nil, requestOptions{})
	nested := requestFixture(t, server, http.MethodGet, "/library/a/b", nil, requestOptions{})
	partial := requestFixture(t, server, http.MethodGet, "/library", nil, hxRequestOptions())
	stylesheetURL := extractStylesheetURL(t, root.Body)

	return optionalCatchAllFixture{
		Root:          root,
		Nested:        nested,
		Partial:       partial,
		Stylesheet:    requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{}),
		StylesheetURL: stylesheetURL,
	}
}

func loadMethodMatrixFixture(t *testing.T) methodMatrixFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "methodmatrixapp")

	home := requestFixture(t, server, http.MethodGet, "/", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, home.Body)

	return methodMatrixFixture{
		Home:          home,
		Get:           requestFixture(t, server, http.MethodGet, "/api/note/ada", nil, requestOptions{}),
		Head:          requestFixture(t, server, http.MethodHead, "/api/note/ada", nil, requestOptions{}),
		Options:       requestFixture(t, server, http.MethodOptions, "/api/note/ada", nil, requestOptions{}),
		Post:          requestFixture(t, server, http.MethodPost, "/api/note/ada", nil, requestOptions{}),
		Patch:         requestFixture(t, server, http.MethodPatch, "/api/note/ada", nil, requestOptions{}),
		Delete:        requestFixture(t, server, http.MethodDelete, "/api/note/ada", nil, requestOptions{}),
		Put:           requestFixture(t, server, http.MethodPut, "/api/note/ada", nil, requestOptions{}),
		Missing:       requestFixture(t, server, http.MethodGet, "/api/missing", nil, requestOptions{}),
		Stylesheet:    requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{}),
		StylesheetURL: stylesheetURL,
	}
}

func loadPrefixAlwaysFixture(t *testing.T) prefixAlwaysFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "i18nprefixalwaysapp")
	opts := secureHostOptions("prefix.example.test")

	homeEN := requestFixture(t, server, http.MethodGet, "/en", nil, opts)
	stylesheetURL := extractStylesheetURL(t, homeEN.Body)

	return prefixAlwaysFixture{
		RootRedirect:       requestFixture(t, server, http.MethodGet, "/", nil, opts),
		HomeEN:             homeEN,
		HomeDE:             requestFixture(t, server, http.MethodGet, "/de", nil, opts),
		NotFoundEN:         requestFixture(t, server, http.MethodGet, "/en/missing", nil, opts),
		NotFoundDE:         requestFixture(t, server, http.MethodGet, "/de/missing", nil, opts),
		PageLoadNotFoundEN: requestFixture(t, server, http.MethodGet, "/en/fail", nil, opts),
		PageLoadNotFoundDE: requestFixture(t, server, http.MethodGet, "/de/fail", nil, opts),
		HelpNotFoundEN:     requestFixture(t, server, http.MethodGet, "/en/help/fail", nil, opts),
		HelpNotFoundDE:     requestFixture(t, server, http.MethodGet, "/de/help/fail", nil, opts),
		GreetRedirect:      requestFixture(t, server, http.MethodGet, "/greet/ada", nil, opts),
		GreetEN:            requestFixture(t, server, http.MethodGet, "/en/greet/ada", nil, opts),
		GreetDE:            requestFixture(t, server, http.MethodGet, "/de/greet/ada", nil, opts),
		GreetPartial: requestFixture(
			t,
			server,
			http.MethodGet,
			"/de/greet/ada",
			nil,
			mergeOptions(opts, hxRequestOptions()),
		),
		Stylesheet:    requestFixture(t, server, http.MethodGet, stylesheetURL, nil, opts),
		StylesheetURL: stylesheetURL,
	}
}

func loadCustomRuntimeFixture(t *testing.T) customRuntimeFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "customruntimeapp")

	home := requestFixture(t, server, http.MethodGet, "/", nil, requestOptions{})
	siteCSSURL := extractSiteCSSURL(t, home.Body)
	templCSSURL := extractStylesheetURL(t, home.Body)

	return customRuntimeFixture{
		Home:          home,
		Extra:         requestFixture(t, server, http.MethodGet, "/debug/ping", nil, requestOptions{}),
		Public:        requestFixture(t, server, http.MethodGet, "/icon.txt", nil, requestOptions{}),
		Health:        requestFixture(t, server, http.MethodGet, "/up", nil, requestOptions{}),
		DefaultHealth: requestFixture(t, server, http.MethodGet, "/healthz", nil, requestOptions{}),
		SiteCSS:       requestFixture(t, server, http.MethodGet, siteCSSURL, nil, requestOptions{}),
		TemplCSS:      requestFixture(t, server, http.MethodGet, templCSSURL, nil, requestOptions{}),
		SiteCSSURL:    siteCSSURL,
		TemplCSSURL:   templCSSURL,
	}
}

func loadTemplRulesFixture(t *testing.T) templRulesFixture {
	t.Helper()

	_, server := startPreparedFixture(t, "templrulesapp")

	card := requestFixture(t, server, http.MethodGet, "/card", nil, requestOptions{})
	stylesheetURL := extractStylesheetURL(t, card.Body)

	return templRulesFixture{
		BaseURL:          server.BaseURL,
		Card:             card,
		Panel:            requestFixture(t, server, http.MethodGet, "/panel", nil, requestOptions{}),
		Board:            requestFixture(t, server, http.MethodGet, "/board", nil, requestOptions{}),
		Deps:             requestFixture(t, server, http.MethodGet, "/deps", nil, requestOptions{}),
		Hooks:            requestFixture(t, server, http.MethodGet, "/hooks", nil, requestOptions{}),
		Vars:             requestFixture(t, server, http.MethodGet, "/vars", nil, requestOptions{}),
		Fallback:         requestFixture(t, server, http.MethodGet, "/fallback", nil, requestOptions{}),
		Metadata:         requestFixture(t, server, http.MethodGet, "/metadata", nil, requestOptions{}),
		MetadataPartial:  requestFixture(t, server, http.MethodGet, "/metadata", nil, hxRequestOptions()),
		Dashboard:        requestFixture(t, server, http.MethodGet, "/dashboard", nil, requestOptions{}),
		DashboardPartial: requestFixture(t, server, http.MethodGet, "/dashboard", nil, hxRequestOptions()),
		TemplCSS:         requestFixture(t, server, http.MethodGet, stylesheetURL, nil, requestOptions{}),
		MeterScript:      requestFixture(t, server, http.MethodGet, "/meter.js", nil, requestOptions{}),
		Stream:           requestStreamFixture(t, server, "/stream", requestOptions{}, "/__e2e/release-stream"),
		StylesheetURL:    stylesheetURL,
	}
}

func secureHostOptions(host string) requestOptions {
	return requestOptions{
		Host: host,
		Headers: map[string]string{
			"X-Forwarded-Proto": "https",
		},
	}
}

func hxRequestOptions() requestOptions {
	return requestOptions{
		Headers: map[string]string{
			"HX-Request": "true",
		},
	}
}

func mergeOptions(base requestOptions, extra requestOptions) requestOptions {
	merged := requestOptions{
		Host:    base.Host,
		Headers: make(map[string]string, len(base.Headers)+len(extra.Headers)),
	}
	for key, value := range base.Headers {
		merged.Headers[key] = value
	}
	for key, value := range extra.Headers {
		merged.Headers[key] = value
	}
	if extra.Host != "" {
		merged.Host = extra.Host
	}
	return merged
}

func extractStylesheetURL(t *testing.T, html string) string {
	t.Helper()

	matches := stylesheetPattern.FindStringSubmatch(html)
	require.Len(t, matches, 2, "stylesheet href not found")
	return normalizeFixtureURL(matches[1])
}

func extractSiteCSSURL(t *testing.T, html string) string {
	t.Helper()

	matches := siteCSSPattern.FindStringSubmatch(html)
	require.Len(t, matches, 2, "site css href not found")
	return normalizeFixtureURL(matches[1])
}

func extractExpensiveCount(t *testing.T, html string) int {
	t.Helper()

	matches := expensiveCountPattern.FindStringSubmatch(html)
	require.Len(t, matches, 2, "expensive count not found")

	value, err := strconv.Atoi(strings.TrimSpace(matches[1]))
	require.NoError(t, err)
	return value
}

func normalizeFixtureURL(value string) string {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		schemeIndex := strings.Index(value, "://")
		if schemeIndex < 0 {
			return value
		}

		hostAndPath := value[schemeIndex+3:]
		if slashIndex := strings.Index(hostAndPath, "/"); slashIndex >= 0 {
			return hostAndPath[slashIndex:]
		}
		return "/"
	}
	return value
}
