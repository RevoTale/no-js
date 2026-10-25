package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkGzipResponse(b *testing.B) {
	handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("response payload ", 32))
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}
}

func TestGzipResponseAllocationBudget(t *testing.T) {
	result := testing.Benchmark(BenchmarkGzipResponse)
	require.Less(t, result.AllocedBytesPerOp(), int64(64*1024), "gzip should reuse compression buffers")
}

func TestGzipResponseIndependentStreams(t *testing.T) {
	handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.URL.Path)
		w.(http.Flusher).Flush()
		_, _ = io.WriteString(w, ":done")
	}))
	for _, path := range []string{"/first", "/second", "/third"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		require.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
		require.Equal(t, path+":done", ungzipBody(t, recorder.Body.Bytes()))
	}
}

func TestGzipResponseConcurrentStreams(t *testing.T) {
	handler := withGzipCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.URL.Path)
	}))
	for _, path := range []string{"/one", "/two", "/three", "/four", "/five", "/six", "/seven", "/eight"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("Accept-Encoding", "gzip")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			require.Equal(t, path, ungzipBody(t, recorder.Body.Bytes()))
		})
	}
}

type gzipFailureWriter struct{ header http.Header }

func (w gzipFailureWriter) Header() http.Header     { return w.header }
func (gzipFailureWriter) WriteHeader(int)           {}
func (gzipFailureWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestGzipResponseRecoversAfterWriteFailure(t *testing.T) {
	writer := &gzipResponseWriter{
		ResponseWriter: gzipFailureWriter{header: make(http.Header)}, compress: true,
	}
	_, _ = writer.Write([]byte("failed response"))
	require.ErrorIs(t, writer.Close(), io.ErrClosedPipe)
	require.NoError(t, writer.Close())
	recorder := httptest.NewRecorder()
	writer = &gzipResponseWriter{ResponseWriter: recorder, compress: true}
	_, err := writer.Write([]byte("next response"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.Equal(t, "next response", ungzipBody(t, recorder.Body.Bytes()))
}
