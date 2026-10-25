package httpserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RevoTale/no-js/v2/framework/staticassets"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func BenchmarkHealthWithGlobalCSS(b *testing.B) {
	classes := make([]templ.CSSClass, 500)
	for index := range classes {
		classes[index] = templ.ComponentCSSClass{ID: fmt.Sprintf("style_%d", index), Class: ".style{}"}
	}
	handler, err := New(Config[struct{}]{
		TemplCSS: &TemplCSSConfig{Manifest: staticassets.Manifest{Hash: "benchmark"}, Classes: classes},
	})
	require.NoError(b, err)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func TestHealthAllocationIndependentOfGlobalCSS(t *testing.T) {
	result := testing.Benchmark(BenchmarkHealthWithGlobalCSS)
	require.Less(t, result.AllocedBytesPerOp(), int64(4096), "health must bypass page CSS state")
}
