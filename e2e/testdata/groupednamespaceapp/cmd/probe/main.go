package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"

	gen "example.com/no-js-e2e/groupednamespaceapp/web/generated"
	runtime "example.com/no-js-e2e/groupednamespaceapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
}

type report struct {
	Notes         responseSnapshot `json:"notes"`
	NotesPartial  responseSnapshot `json:"notes_partial"`
	Guides        responseSnapshot `json:"guides"`
	Tags          responseSnapshot `json:"tags"`
	TagsPartial   responseSnapshot `json:"tags_partial"`
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

	notes := request(handler, http.MethodGet, "/discover/notes", nil)
	stylesheetURL := stylesheetURLFromBody(notes.Body)

	out := report{
		Notes:         notes,
		NotesPartial:  request(handler, http.MethodGet, "/discover/notes", map[string]string{"HX-Request": "true"}),
		Guides:        request(handler, http.MethodGet, "/discover/guides", nil),
		Tags:          request(handler, http.MethodGet, "/discover/tags", nil),
		TagsPartial:   request(handler, http.MethodGet, "/discover/tags", map[string]string{"HX-Request": "true"}),
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
