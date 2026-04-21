package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"

	gen "example.com/no-js-e2e/customruntimeapp/web/generated"
	runtime "example.com/no-js-e2e/customruntimeapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
	Location             string `json:"location"`
	XMainMiddleware      string `json:"x_main_middleware"`
}

type report struct {
	Home          responseSnapshot `json:"home"`
	Extra         responseSnapshot `json:"extra"`
	Public        responseSnapshot `json:"public"`
	Health        responseSnapshot `json:"health"`
	DefaultHealth responseSnapshot `json:"default_health"`
	SiteCSS       responseSnapshot `json:"site_css"`
	TemplCSS      responseSnapshot `json:"templ_css"`
	SiteCSSURL    string           `json:"site_css_url"`
	TemplCSSURL   string           `json:"templ_css_url"`
}

var siteCSSPattern = regexp.MustCompile(`id="site-css"[^>]*href="([^"]+/site\.css)"`)
var templCSSPattern = regexp.MustCompile(`href="([^"]+/styles/templ\.css)"`)

func main() {
	appContext := &runtime.Context{}
	bundle := gen.Bundle(appContext)
	bundle.TemplCSSClasses = gen.TemplCSSClasses

	mainMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Main-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
		App: bundle,
		Custom: httpserver.CustomConfig{
			MainMiddlewares: []func(http.Handler) http.Handler{
				mainMiddleware,
			},
			ExtraRoutes: func(mux *http.ServeMux) error {
				mux.HandleFunc("/debug/ping", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("extra"))
				})
				return nil
			},
			StaticAssets: &httpserver.StaticAssetsConfig{
				ManifestPath: "web/assets-build/manifest.json",
				URLPrefix:    "/build/",
			},
			PublicFiles: &httpserver.PublicFilesConfig{
				Dir: "web/custom-public",
			},
			HealthPath: "/up",
			HealthBody: "alive",
		},
	})
	if err != nil {
		panic(err)
	}

	home := request(handler, http.MethodGet, "/", nil)
	siteCSSURL := extractURL(siteCSSPattern, home.Body, "site css")
	templCSSURL := extractURL(templCSSPattern, home.Body, "templ css")

	out := report{
		Home:          home,
		Extra:         request(handler, http.MethodGet, "/debug/ping", nil),
		Public:        request(handler, http.MethodGet, "/icon.txt", nil),
		Health:        request(handler, http.MethodGet, "/up", nil),
		DefaultHealth: request(handler, http.MethodGet, "/healthz", nil),
		SiteCSS:       request(handler, http.MethodGet, siteCSSURL, nil),
		TemplCSS:      request(handler, http.MethodGet, templCSSURL, nil),
		SiteCSSURL:    siteCSSURL,
		TemplCSSURL:   templCSSURL,
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}

func request(handler http.Handler, method string, path string, headers map[string]string) responseSnapshot {
	req := httptest.NewRequest(method, path, nil)
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
		XMainMiddleware:      rec.Header().Get("X-Main-Middleware"),
	}
}

func extractURL(pattern *regexp.Regexp, body string, label string) string {
	matches := pattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		panic(fmt.Errorf("%s link missing from body", label))
	}
	return matches[1]
}
