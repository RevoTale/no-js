package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"

	gen "example.com/templcssapp/web/generated"
	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

const requestBaseURL = "https://docs.example.test"

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
	Location             string `json:"location"`
}

type report struct {
	AuthorDE                  responseSnapshot `json:"author_de"`
	AuthorEN                  responseSnapshot `json:"author_en"`
	AuthorENRedirect          responseSnapshot `json:"author_en_redirect"`
	AuthorMissing             responseSnapshot `json:"author_missing"`
	AuthorError               responseSnapshot `json:"author_error"`
	Dashboard                 responseSnapshot `json:"dashboard"`
	DashboardPartial          responseSnapshot `json:"dashboard_partial"`
	PingDE                    responseSnapshot `json:"ping_de"`
	Robots                    responseSnapshot `json:"robots"`
	Feed                      responseSnapshot `json:"feed"`
	AuthorFeed                responseSnapshot `json:"author_feed"`
	Sitemap                   responseSnapshot `json:"sitemap"`
	SitemapIndex              responseSnapshot `json:"sitemap_index"`
	SitemapChunk              responseSnapshot `json:"sitemap_chunk"`
	Favicon                   responseSnapshot `json:"favicon"`
	SiteCSS                   responseSnapshot `json:"site_css"`
	TemplCSS                  responseSnapshot `json:"templ_css"`
	Health                    responseSnapshot `json:"health"`
	SiteCSSURL                string           `json:"site_css_url"`
	TemplCSSURL               string           `json:"templ_css_url"`
	ExpensiveLoadsAfterFirst  int              `json:"expensive_loads_after_first"`
	ExpensiveLoadsAfterSecond int              `json:"expensive_loads_after_second"`
}

var siteCSSPattern = regexp.MustCompile(`id="site-css"[^>]*href="([^"]+/site\.css)"`)
var templCSSPattern = regexp.MustCompile(`href="([^"]+/styles/templ\.css)"`)

func main() {
	appContext := runtime.NewContext()
	bundle := gen.Bundle(appContext)
	bundle.TemplCSSClasses = gen.TemplCSSClasses

	handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
		App: bundle,
		Custom: httpserver.CustomConfig{
			LogServerError: func(error) {},
		},
	})
	if err != nil {
		panic(err)
	}

	authorDE := request(handler, http.MethodGet, "/de/author/ada", nil)
	siteCSSURL := extractURL(siteCSSPattern, authorDE.Body, "site css")
	templCSSURL := extractURL(templCSSPattern, authorDE.Body, "templ css")
	expensiveLoadsAfterFirst := appContext.ExpensiveLoadCount()

	authorENRedirect := request(handler, http.MethodGet, "/en/author/ada", nil)
	authorEN := request(handler, http.MethodGet, "/author/ada", nil)
	expensiveLoadsAfterSecond := appContext.ExpensiveLoadCount()

	out := report{
		AuthorDE:                  authorDE,
		AuthorEN:                  authorEN,
		AuthorENRedirect:          authorENRedirect,
		AuthorMissing:             request(handler, http.MethodGet, "/de/author/missing", nil),
		AuthorError:               request(handler, http.MethodGet, "/de/author/boom", nil),
		Dashboard:                 request(handler, http.MethodGet, "/de/dashboard", nil),
		DashboardPartial:          request(handler, http.MethodGet, "/de/dashboard", map[string]string{"HX-Request": "true"}),
		PingDE:                    request(handler, http.MethodGet, "/de/api/ping", nil),
		Robots:                    request(handler, http.MethodGet, "/robots.txt", nil),
		Feed:                      request(handler, http.MethodGet, "/feed.xml", nil),
		AuthorFeed:                request(handler, http.MethodGet, "/author/ada/feed.xml", nil),
		Sitemap:                   request(handler, http.MethodGet, "/sitemap.xml", nil),
		SitemapIndex:              request(handler, http.MethodGet, "/sitemap-index.xml", nil),
		SitemapChunk:              request(handler, http.MethodGet, "/sitemap/authors.xml", nil),
		Favicon:                   request(handler, http.MethodGet, "/favicon.ico", nil),
		SiteCSS:                   request(handler, http.MethodGet, siteCSSURL, nil),
		TemplCSS:                  request(handler, http.MethodGet, templCSSURL, nil),
		Health:                    request(handler, http.MethodGet, "/healthz", nil),
		SiteCSSURL:                siteCSSURL,
		TemplCSSURL:               templCSSURL,
		ExpensiveLoadsAfterFirst:  expensiveLoadsAfterFirst,
		ExpensiveLoadsAfterSecond: expensiveLoadsAfterSecond,
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}

func request(handler http.Handler, method string, target string, headers map[string]string) responseSnapshot {
	if strings.HasPrefix(strings.TrimSpace(target), "/") {
		target = requestBaseURL + target
	}

	req := httptest.NewRequest(method, target, nil)
	req.TLS = &tls.ConnectionState{}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return responseSnapshot{
		Status:               rec.Code,
		Body:                 rec.Body.String(),
		ContentType:          rec.Header().Get("Content-Type"),
		HXTriggerAfterSettle: rec.Header().Get("HX-Trigger-After-Settle"),
		Location:             rec.Header().Get("Location"),
	}
}

func extractURL(pattern *regexp.Regexp, body string, label string) string {
	matches := pattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		panic(fmt.Errorf("%s link missing from body", label))
	}
	return matches[1]
}
