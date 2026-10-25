package httpserver

import (
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlreadyCompressedAssetsUseIdentityEncoding(t *testing.T) {
	for _, contentType := range []string{
		"image/png", "image/jpeg", "image/webp", "font/woff2", "application/zip", "video/mp4",
	} {
		t.Run(contentType, func(t *testing.T) {
			body := []byte("already compressed payload")
			handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Content-Length", strconv.Itoa(len(body)))
				_, _ = w.Write(body)
			}))
			request := httptest.NewRequest(http.MethodGet, "/asset", nil)
			request.Header.Set("Accept-Encoding", "gzip")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			require.Empty(t, recorder.Header().Get("Content-Encoding"))
			require.Equal(t, body, recorder.Body.Bytes())
			require.Equal(t, strconv.Itoa(len(body)), recorder.Header().Get("Content-Length"))
		})
	}
}

func BenchmarkCompressedAsset(b *testing.B) {
	body := make([]byte, 32768)
	random := rand.New(rand.NewPCG(1, 2))
	for index := range body {
		body[index] = byte(random.Uint32())
	}
	handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(body)
	}))
	request := httptest.NewRequest(http.MethodGet, "/asset.png", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func TestSVGStillUsesGzip(t *testing.T) {
	body := []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		_, _ = w.Write(body)
	}))
	request := httptest.NewRequest(http.MethodGet, "/asset.svg", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	require.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
	require.Equal(t, string(body), ungzipBody(t, recorder.Body.Bytes()))
}
