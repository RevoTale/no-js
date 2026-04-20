package stream

import (
	"net/http"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtimeCtx framework.RuntimeContext[*runtime.Context],
	w http.ResponseWriter,
	r *http.Request,
	params StreamParams,
) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := w.Write([]byte("first")); err != nil {
		return err
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	flusher.Flush()

	runtimeCtx.AppContext().StreamState().Wait()

	_, err := w.Write([]byte("second"))
	return err
}
