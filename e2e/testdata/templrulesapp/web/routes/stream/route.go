package stream

import (
	"net/http"

	"example.com/no-js-e2e/templrulesapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtimeCtx framework.RuntimeContext[*view.Context],
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
