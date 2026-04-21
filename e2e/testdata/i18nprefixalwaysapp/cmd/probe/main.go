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

	gen "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated"
	runtime "example.com/no-js-e2e/i18nprefixalwaysapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

const requestBaseURL = "https://prefix.example.test"

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
	Location             string `json:"location"`
}

type report struct {
	RootRedirect  responseSnapshot `json:"root_redirect"`
	HomeEN        responseSnapshot `json:"home_en"`
	HomeDE        responseSnapshot `json:"home_de"`
	GreetRedirect responseSnapshot `json:"greet_redirect"`
	GreetEN       responseSnapshot `json:"greet_en"`
	GreetDE       responseSnapshot `json:"greet_de"`
	GreetPartial  responseSnapshot `json:"greet_partial"`
	Stylesheet    responseSnapshot `json:"stylesheet"`
	StylesheetURL string           `json:"stylesheet_url"`
}

var stylesheetPattern = regexp.MustCompile(`href="([^"]+/styles/templ\.css)"`)

func main() {
	appContext := runtime.NewContext()
	bundle := gen.Bundle(appContext)
	bundle.TemplCSSClasses = gen.TemplCSSClasses

	handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
		App: bundle,
	})
	if err != nil {
		panic(err)
	}

	greetEN := request(handler, http.MethodGet, "/en/greet/ada", nil)
	stylesheetURL := extractURL(stylesheetPattern, greetEN.Body, "templ css")

	out := report{
		RootRedirect:  request(handler, http.MethodGet, "/", nil),
		HomeEN:        request(handler, http.MethodGet, "/en", nil),
		HomeDE:        request(handler, http.MethodGet, "/de", nil),
		GreetRedirect: request(handler, http.MethodGet, "/greet/ada", nil),
		GreetEN:       greetEN,
		GreetDE:       request(handler, http.MethodGet, "/de/greet/ada", nil),
		GreetPartial:  request(handler, http.MethodGet, "/de/greet/ada", map[string]string{"HX-Request": "true"}),
		Stylesheet:    request(handler, http.MethodGet, stylesheetURL, nil),
		StylesheetURL: stylesheetURL,
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
