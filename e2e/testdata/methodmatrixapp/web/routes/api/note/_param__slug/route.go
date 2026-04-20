package routes

import (
	"encoding/json"
	"net/http"

	view "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiNoteParamSlugParams,
) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(map[string]string{
		"method": http.MethodGet,
		"slug":   params.Slug,
		"path":   r.URL.Path,
	})
}

func POST(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiNoteParamSlugParams,
) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Location", "/api/note/"+params.Slug)
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(map[string]string{
		"method": http.MethodPost,
		"slug":   params.Slug,
	})
}

func PATCH(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiNoteParamSlugParams,
) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	return json.NewEncoder(w).Encode(map[string]string{
		"method": http.MethodPatch,
		"slug":   params.Slug,
	})
}

func DELETE(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiNoteParamSlugParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
