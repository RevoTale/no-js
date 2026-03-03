package framework

import (
	"context"
	"fmt"
	"net/http"

	frameworki18n "blog/framework/i18n"
	"blog/framework/metagen"
	"github.com/a-h/templ"
)

type EmptyParams struct{}

type SlugParams struct {
	Slug string
}

type ParamsParser[P interface{}] func(path string) (P, bool)

type PageLoader[C interface{}, P interface{}, VM interface{}] func(
	ctx context.Context,
	appCtx C,
	r *http.Request,
	params P,
) (VM, error)

type PageMetaGen[C interface{}, P interface{}] func(
	ctx context.Context,
	appCtx C,
	r *http.Request,
	params P,
) (metagen.Metadata, error)

type PageRenderer[VM interface{}] func(view VM) templ.Component

type LayoutRenderer[VM interface{}] func(meta metagen.Metadata, view VM, child templ.Component) templ.Component

type PageModule[C interface{}, P interface{}, VM interface{}] struct {
	Pattern     string
	ParseParams ParamsParser[P]
	MetaGen     PageMetaGen[C, P]
	Load        PageLoader[C, P, VM]
	Render      PageRenderer[VM]
	Layouts     []LayoutRenderer[VM]
}

type RuntimeContext[C interface{}] interface {
	AppContext() C
	IsPartialRequest(r *http.Request) bool
	RenderPage(r *http.Request, w http.ResponseWriter, component templ.Component, meta metagen.Metadata) error
	IsNotFound(err error) bool
	RespondNotFound(w http.ResponseWriter, r *http.Request, notFoundContext NotFoundContext)
	RespondServerError(w http.ResponseWriter, err error)
}

type NotFoundSource string

const (
	NotFoundSourcePageLoad       NotFoundSource = "page_load"
	NotFoundSourceMetaGen        NotFoundSource = "meta_gen"
	NotFoundSourceUnmatchedRoute NotFoundSource = "unmatched_route"
)

type NotFoundContext struct {
	RequestPath         string
	MatchedRoutePattern string
	Locale              string
	Source              NotFoundSource
}

type RouteHandler[C interface{}] interface {
	TryServe(runtime RuntimeContext[C], w http.ResponseWriter, r *http.Request) bool
}

type PageOnlyRouteHandler[C interface{}, P interface{}, VM interface{}] struct {
	Page PageModule[C, P, VM]
}

func (h PageOnlyRouteHandler[C, P, VM]) TryServe(
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
) bool {
	return servePageModule(runtime, w, r, h.Page)
}

func applyLayouts[VM interface{}](
	layouts []LayoutRenderer[VM],
	meta metagen.Metadata,
	view VM,
	child templ.Component,
) templ.Component {
	wrapped := child
	for idx := len(layouts) - 1; idx >= 0; idx-- {
		wrapped = layouts[idx](meta, view, wrapped)
	}
	return wrapped
}

func servePageModule[C interface{}, P interface{}, VM interface{}](
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
	module PageModule[C, P, VM],
) bool {
	params, ok := module.ParseParams(r.URL.Path)
	if !ok {
		return false
	}

	meta := metagen.Metadata{}
	if module.MetaGen != nil {
		var err error
		meta, err = module.MetaGen(r.Context(), runtime.AppContext(), r, params)
		if err != nil {
			handleModuleError(runtime, w, r, err, module.Pattern, NotFoundSourceMetaGen, "meta")
			return true
		}
	}
	meta = metagen.Normalize(meta)

	view, err := module.Load(r.Context(), runtime.AppContext(), r, params)
	if err != nil {
		handleModuleError(runtime, w, r, err, module.Pattern, NotFoundSourcePageLoad, "load")
		return true
	}

	component := module.Render(view)
	if !runtime.IsPartialRequest(r) {
		component = applyLayouts(module.Layouts, meta, view, component)
	}
	if err := runtime.RenderPage(r, w, component, meta); err != nil {
		runtime.RespondServerError(w, fmt.Errorf("render route %q: %w", module.Pattern, err))
	}
	return true
}

func handleModuleError[C interface{}](
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
	err error,
	routePattern string,
	source NotFoundSource,
	stage string,
) {
	if runtime.IsNotFound(err) {
		runtime.RespondNotFound(w, r, NotFoundContext{
			RequestPath:         r.URL.Path,
			MatchedRoutePattern: routePattern,
			Locale:              frameworki18n.LocaleFromContext(r.Context()),
			Source:              source,
		})
		return
	}

	runtime.RespondServerError(w, fmt.Errorf("%s route %q: %w", stage, routePattern, err))
}
