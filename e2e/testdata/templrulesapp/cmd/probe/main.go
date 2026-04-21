package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"

	gen "example.com/no-js-e2e/templrulesapp/web/generated"
	runtime "example.com/no-js-e2e/templrulesapp/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

type responseSnapshot struct {
	Status               int    `json:"status"`
	Body                 string `json:"body"`
	ContentType          string `json:"content_type"`
	HXTriggerAfterSettle string `json:"hx_trigger_after_settle"`
}

type streamSnapshot struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	FirstChunk  string `json:"first_chunk"`
	Body        string `json:"body"`
}

type report struct {
	BaseURL          string           `json:"base_url"`
	Card             responseSnapshot `json:"card"`
	Panel            responseSnapshot `json:"panel"`
	Board            responseSnapshot `json:"board"`
	Deps             responseSnapshot `json:"deps"`
	Hooks            responseSnapshot `json:"hooks"`
	Vars             responseSnapshot `json:"vars"`
	Fallback         responseSnapshot `json:"fallback"`
	Metadata         responseSnapshot `json:"metadata"`
	MetadataPartial  responseSnapshot `json:"metadata_partial"`
	Dashboard        responseSnapshot `json:"dashboard"`
	DashboardPartial responseSnapshot `json:"dashboard_partial"`
	TemplCSS         responseSnapshot `json:"templ_css"`
	MeterScript      responseSnapshot `json:"meter_script"`
	Stream           streamSnapshot   `json:"stream"`
	StylesheetURL    string           `json:"stylesheet_url"`
}

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

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	defer server.Close()

	client := server.Client()
	if transport, ok := client.Transport.(*http.Transport); ok {
		clone := transport.Clone()
		clone.DisableCompression = true
		client.Transport = clone
	}

	card := request(client, server.URL, http.MethodGet, "/card", nil)
	stylesheetURL := extractURL(templCSSPattern, card.Body, "templ css")

	out := report{
		BaseURL:          server.URL,
		Card:             card,
		Panel:            request(client, server.URL, http.MethodGet, "/panel", nil),
		Board:            request(client, server.URL, http.MethodGet, "/board", nil),
		Deps:             request(client, server.URL, http.MethodGet, "/deps", nil),
		Hooks:            request(client, server.URL, http.MethodGet, "/hooks", nil),
		Vars:             request(client, server.URL, http.MethodGet, "/vars", nil),
		Fallback:         request(client, server.URL, http.MethodGet, "/fallback", nil),
		Metadata:         request(client, server.URL, http.MethodGet, "/metadata", nil),
		MetadataPartial:  request(client, server.URL, http.MethodGet, "/metadata", map[string]string{"HX-Request": "true"}),
		Dashboard:        request(client, server.URL, http.MethodGet, "/dashboard", nil),
		DashboardPartial: request(client, server.URL, http.MethodGet, "/dashboard", map[string]string{"HX-Request": "true"}),
		TemplCSS:         request(client, server.URL, http.MethodGet, stylesheetURL, nil),
		MeterScript:      request(client, server.URL, http.MethodGet, "/meter.js", nil),
		Stream:           requestStream(client, server.URL, appContext),
		StylesheetURL:    stylesheetURL,
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}

func request(client *http.Client, baseURL string, method string, target string, headers map[string]string) responseSnapshot {
	req, err := http.NewRequest(method, normalizeURL(baseURL, target), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "identity")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return responseSnapshot{
		Status:               resp.StatusCode,
		Body:                 string(body),
		ContentType:          resp.Header.Get("Content-Type"),
		HXTriggerAfterSettle: resp.Header.Get("HX-Trigger-After-Settle"),
	}
}

func requestStream(client *http.Client, baseURL string, appContext *runtime.Context) streamSnapshot {
	req, err := http.NewRequest(http.MethodGet, normalizeURL(baseURL, "/stream"), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	firstChunk := make([]byte, len("first"))
	if _, err := io.ReadFull(resp.Body, firstChunk); err != nil {
		panic(err)
	}

	appContext.ReleaseStream()

	rest, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return streamSnapshot{
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		FirstChunk:  string(firstChunk),
		Body:        string(firstChunk) + string(rest),
	}
}

func normalizeURL(baseURL string, target string) string {
	trimmed := strings.TrimSpace(target)
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(trimmed, "/")
}

func extractURL(pattern *regexp.Regexp, body string, label string) string {
	matches := pattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		panic(fmt.Errorf("%s link missing from body", label))
	}
	return matches[1]
}
