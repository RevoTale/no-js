package routes

import (
	"encoding/json"
	"net/http"

	view "example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
)

func GET(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiPingParams,
) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(map[string]string{
		"locale": frameworki18n.LocaleFromContext(r.Context()),
		"path":   r.URL.Path,
	})
}
