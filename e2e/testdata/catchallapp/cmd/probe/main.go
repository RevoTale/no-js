package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"

	gen "example.com/templcssapp/web/generated"
	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
}

type report struct {
	Docs          responseSnapshot `json:"docs"`
	Nested        responseSnapshot `json:"nested"`
	Partial       responseSnapshot `json:"partial"`
	Missing       responseSnapshot `json:"missing"`
	Stylesheet    responseSnapshot `json:"stylesheet"`
	StylesheetURL string           `json:"stylesheet_url"`
}

var stylesheetPattern = regexp.MustCompile(`href="([^"]+/styles/templ\.css)"`)

func main() {
	appContext := &runtime.Context{}
	bundle := gen.Bundle(appContext)
	bundle.TemplCSSClasses = gen.TemplCSSClasses

	handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
		App: bundle,
	})
	if err != nil {
		panic(err)
	}

	docs := request(handler, http.MethodGet, "/docs/a/b", nil)
	stylesheetURL := stylesheetURLFromBody(docs.Body)

	out := report{
		Docs:          docs,
		Nested:        request(handler, http.MethodGet, "/docs/alpha/beta/gamma", nil),
		Partial:       request(handler, http.MethodGet, "/docs/a/b", map[string]string{"HX-Request": "true"}),
		Missing:       request(handler, http.MethodGet, "/docs", nil),
		Stylesheet:    request(handler, http.MethodGet, stylesheetURL, nil),
		StylesheetURL: stylesheetURL,
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
	}
}

func stylesheetURLFromBody(body string) string {
	matches := stylesheetPattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		panic(fmt.Errorf("stylesheet link missing from body"))
	}
	return matches[1]
}
