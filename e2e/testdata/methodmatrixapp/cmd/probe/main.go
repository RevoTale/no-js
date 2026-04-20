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
	Location             string `json:"location"`
	Allow                string `json:"allow"`
}

type report struct {
	Home          responseSnapshot `json:"home"`
	Get           responseSnapshot `json:"get"`
	Head          responseSnapshot `json:"head"`
	Options       responseSnapshot `json:"options"`
	Post          responseSnapshot `json:"post"`
	Patch         responseSnapshot `json:"patch"`
	Delete        responseSnapshot `json:"delete"`
	Put           responseSnapshot `json:"put"`
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

	home := request(handler, http.MethodGet, "/", nil)
	stylesheetURL := stylesheetURLFromBody(home.Body)

	out := report{
		Home:          home,
		Get:           request(handler, http.MethodGet, "/api/note/ada", nil),
		Head:          request(handler, http.MethodHead, "/api/note/ada", nil),
		Options:       request(handler, http.MethodOptions, "/api/note/ada", nil),
		Post:          request(handler, http.MethodPost, "/api/note/ada", nil),
		Patch:         request(handler, http.MethodPatch, "/api/note/ada", nil),
		Delete:        request(handler, http.MethodDelete, "/api/note/ada", nil),
		Put:           request(handler, http.MethodPut, "/api/note/ada", nil),
		Missing:       request(handler, http.MethodGet, "/api/missing", nil),
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
		Location:             rec.Header().Get("Location"),
		Allow:                rec.Header().Get("Allow"),
	}
}

func stylesheetURLFromBody(body string) string {
	matches := stylesheetPattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		panic(fmt.Errorf("stylesheet link missing from body"))
	}
	return matches[1]
}
