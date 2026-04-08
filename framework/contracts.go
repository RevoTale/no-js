package framework

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
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

type PageMetaGenContext[C interface{}, P interface{}] func(
	meta MetaContext[C],
	params P,
) (metagen.Metadata, error)

type PageRenderer[VM interface{}] func(view VM) templ.Component

type LayoutRenderer[VM interface{}] func(meta metagen.Metadata, view VM, child templ.Component) templ.Component

type PageComposer[C interface{}, P interface{}, VM interface{}] func(
	ctx context.Context,
	runtime RuntimeContext[C],
	r *http.Request,
	meta metagen.Metadata,
	view VM,
	params P,
	partial bool,
) (templ.Component, error)

type MethodRouteAction[C interface{}, P interface{}] func(
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
	params P,
) error

type PageModule[C interface{}, P interface{}, VM interface{}] struct {
	RouteID             string
	Pattern             string
	ParseParams         ParamsParser[P]
	MetaGen             PageMetaGen[C, P]
	MetaGenContext      PageMetaGenContext[C, P]
	MetaGenName         string
	MetaGenChain        []PageMetaGen[C, P]
	MetaGenContextChain []PageMetaGenContext[C, P]
	MetaGenChainNames   []string
	Load                PageLoader[C, P, VM]
	LoadName            string
	Compose             PageComposer[C, P, VM]
	Render              PageRenderer[VM]
	Layouts             []LayoutRenderer[VM]
	RootLayout          func(meta metagen.Metadata, locale string, child templ.Component) templ.Component
	ErrorPage           func(appCtx C, r *http.Request) templ.Component
}

type MethodRouteModule[C interface{}, P interface{}] struct {
	RouteID     string
	Pattern     string
	ParseParams ParamsParser[P]
	GET         MethodRouteAction[C, P]
	POST        MethodRouteAction[C, P]
	PUT         MethodRouteAction[C, P]
	PATCH       MethodRouteAction[C, P]
	DELETE      MethodRouteAction[C, P]
	HEAD        MethodRouteAction[C, P]
	OPTIONS     MethodRouteAction[C, P]
}

type RuntimeContext[C interface{}] interface {
	AppContext() C
	I18n() *frameworki18n.Resolver
	ResolveRoot(r *http.Request) *url.URL
	IsPartialRequest(r *http.Request) bool
	RenderPage(r *http.Request, w http.ResponseWriter, component templ.Component, meta metagen.Metadata) error
	RespondNotFound(w http.ResponseWriter, r *http.Request, notFoundContext NotFoundContext)
	RespondServerError(w http.ResponseWriter, err error)
	LogServerError(err error)
	LogResolverTiming(event ResolverTiming)
}

type ResolverStage string

const (
	ResolverStageMetaGen ResolverStage = "meta_gen"
	ResolverStageLoad    ResolverStage = "load"
)

type ResolverTiming struct {
	RoutePattern string
	Stage        ResolverStage
	Method       string
	Duration     time.Duration
	Err          error
}

type NotFoundSource string

const (
	NotFoundSourcePageLoad       NotFoundSource = "page_load"
	NotFoundSourceMetaGen        NotFoundSource = "meta_gen"
	NotFoundSourceUnmatchedRoute NotFoundSource = "unmatched_route"
)

type NotFoundContext struct {
	RequestPath         string
	MatchedRouteID      string
	MatchedRoutePattern string
	Locale              string
	Source              NotFoundSource
}

type RouteHandler[C interface{}] interface {
	TryServe(runtime RuntimeContext[C], w http.ResponseWriter, r *http.Request) bool
}

type PathMatcher interface {
	MatchPath(path string) bool
}

type PageOnlyRouteHandler[C interface{}, P interface{}, VM interface{}] struct {
	Page PageModule[C, P, VM]
}

type MethodOnlyRouteHandler[C interface{}, P interface{}] struct {
	Route MethodRouteModule[C, P]
}

func (h PageOnlyRouteHandler[C, P, VM]) MatchPath(path string) bool {
	_, ok := h.Page.ParseParams(path)
	return ok
}

func (h PageOnlyRouteHandler[C, P, VM]) TryServe(
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
) bool {
	return servePageModule(runtime, w, r, h.Page)
}

func (h MethodOnlyRouteHandler[C, P]) MatchPath(path string) bool {
	_, ok := h.Route.ParseParams(path)
	return ok
}

func (h MethodOnlyRouteHandler[C, P]) TryServe(
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
) bool {
	return serveMethodRoute(runtime, w, r, h.Route)
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

	type metadataResult struct {
		meta metagen.Metadata
		err  error
	}
	type pageLoadResult struct {
		view VM
		err  error
	}

	ctx, cancel := context.WithCancel(r.Context())
	ctx = WithRequestCache(ctx)
	defer cancel()
	appCtx := runtime.AppContext()

	metaCh := make(chan metadataResult, 1)
	go func() {
		meta, err := resolveMetadata(runtime, ctx, appCtx, r, params, module)
		metaCh <- metadataResult{meta: meta, err: err}
	}()

	loadCh := make(chan pageLoadResult, 1)
	go func() {
		startedAt := time.Now()
		view, err := module.Load(ctx, appCtx, r, params)
		runtime.LogResolverTiming(ResolverTiming{
			RoutePattern: module.Pattern,
			Stage:        ResolverStageLoad,
			Method:       loadMethodName(module),
			Duration:     time.Since(startedAt),
			Err:          err,
		})
		loadCh <- pageLoadResult{view: view, err: err}
	}()

	metaResult := <-metaCh
	if metaResult.err != nil {
		handleModuleError(runtime, w, r, metaResult.err, module.RouteID, module.Pattern, NotFoundSourceMetaGen, "meta")
		return true
	}
	meta := metagen.Normalize(metaResult.meta)

	var loadOnce sync.Once
	var loadResult pageLoadResult
	awaitLoad := func() pageLoadResult {
		loadOnce.Do(func() {
			loadResult = <-loadCh
		})
		return loadResult
	}

	partial := runtime.IsPartialRequest(r)
	if partial || module.RootLayout == nil {
		result := awaitLoad()
		if result.err != nil {
			handleModuleError(runtime, w, r, result.err, module.RouteID, module.Pattern, NotFoundSourcePageLoad, "load")
			return true
		}

		component, err := composePageComponent(ctx, runtime, r, meta, result.view, params, partial, module)
		if err != nil {
			handleModuleError(runtime, w, r, err, module.RouteID, module.Pattern, NotFoundSourcePageLoad, "compose")
			return true
		}
		if err := runtime.RenderPage(r, w, component, meta); err != nil {
			runtime.RespondServerError(w, fmt.Errorf("render route %q: %w", module.Pattern, err))
		}
		return true
	}

	locale := frameworki18n.LocaleFromContext(r.Context())
	streamedBody := templ.ComponentFunc(func(renderCtx context.Context, writer io.Writer) error {
		result := awaitLoad()
		if result.err != nil {
			runtime.LogServerError(fmt.Errorf("load route %q after stream start: %w", module.Pattern, result.err))
			if module.ErrorPage == nil {
				return nil
			}
			errorComponent := module.ErrorPage(appCtx, r)
			if errorComponent == nil {
				return nil
			}
			return errorComponent.Render(renderCtx, writer)
		}

		component, err := composePageComponent(ctx, runtime, r, meta, result.view, params, false, module)
		if err != nil {
			return fmt.Errorf("compose route %q: %w", module.Pattern, err)
		}
		return component.Render(renderCtx, writer)
	})

	component := module.RootLayout(meta, locale, streamedBody)
	if component == nil {
		component = streamedBody
	}
	if err := runtime.RenderPage(r, w, component, meta); err != nil {
		runtime.RespondServerError(w, fmt.Errorf("render route %q: %w", module.Pattern, err))
	}
	return true
}

func composePageComponent[C interface{}, P interface{}, VM interface{}](
	ctx context.Context,
	runtime RuntimeContext[C],
	r *http.Request,
	meta metagen.Metadata,
	view VM,
	params P,
	partial bool,
	module PageModule[C, P, VM],
) (templ.Component, error) {
	if module.Compose != nil {
		return module.Compose(ctx, runtime, r, meta, view, params, partial)
	}

	component := module.Render(view)
	if !partial {
		component = applyLayouts(module.Layouts, meta, view, component)
	}
	return component, nil
}

func serveMethodRoute[C interface{}, P interface{}](
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
	module MethodRouteModule[C, P],
) bool {
	params, ok := module.ParseParams(r.URL.Path)
	if !ok {
		return false
	}

	action, methodName, allow := methodRouteAction(module, r.Method)
	if action == nil {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return true
	}

	if r.Method == http.MethodHead && module.HEAD == nil && module.GET != nil {
		w = headResponseWriter{ResponseWriter: w}
	}

	if err := action(runtime, w, r, params); err != nil {
		handleModuleError(
			runtime,
			w,
			r,
			err,
			module.RouteID,
			module.Pattern,
			NotFoundSourcePageLoad,
			strings.ToLower(methodName),
		)
	}
	return true
}

func methodRouteAction[C interface{}, P interface{}](
	module MethodRouteModule[C, P],
	method string,
) (MethodRouteAction[C, P], string, []string) {
	allowed := allowedMethods(module)
	switch method {
	case http.MethodGet:
		return module.GET, http.MethodGet, allowed
	case http.MethodPost:
		return module.POST, http.MethodPost, allowed
	case http.MethodPut:
		return module.PUT, http.MethodPut, allowed
	case http.MethodPatch:
		return module.PATCH, http.MethodPatch, allowed
	case http.MethodDelete:
		return module.DELETE, http.MethodDelete, allowed
	case http.MethodHead:
		if module.HEAD != nil {
			return module.HEAD, http.MethodHead, allowed
		}
		if module.GET != nil {
			return module.GET, http.MethodGet, allowed
		}
		return nil, http.MethodHead, allowed
	case http.MethodOptions:
		return module.OPTIONS, http.MethodOptions, allowed
	default:
		return nil, method, allowed
	}
}

func allowedMethods[C interface{}, P interface{}](module MethodRouteModule[C, P]) []string {
	methods := make([]string, 0, 7)
	if module.GET != nil {
		methods = append(methods, http.MethodGet)
	}
	if module.POST != nil {
		methods = append(methods, http.MethodPost)
	}
	if module.PUT != nil {
		methods = append(methods, http.MethodPut)
	}
	if module.PATCH != nil {
		methods = append(methods, http.MethodPatch)
	}
	if module.DELETE != nil {
		methods = append(methods, http.MethodDelete)
	}
	if module.HEAD != nil || module.GET != nil {
		methods = append(methods, http.MethodHead)
	}
	if module.OPTIONS != nil || len(methods) > 0 {
		methods = append(methods, http.MethodOptions)
	}
	return methods
}

type headResponseWriter struct {
	http.ResponseWriter
}

func (writer headResponseWriter) Write(body []byte) (int, error) {
	return len(body), nil
}

func resolveMetadata[C interface{}, P interface{}, VM interface{}](
	runtime RuntimeContext[C],
	ctx context.Context,
	appCtx C,
	r *http.Request,
	params P,
	module PageModule[C, P, VM],
) (metagen.Metadata, error) {
	contextChain := module.MetaGenContextChain
	if len(contextChain) == 0 && module.MetaGenContext != nil {
		contextChain = append(contextChain, module.MetaGenContext)
	}
	if len(contextChain) > 0 {
		return resolveMetadataWithContext(
			runtime,
			ctx,
			appCtx,
			r,
			params,
			module.Pattern,
			contextChain,
			module.MetaGenChainNames,
		)
	}

	chain := module.MetaGenChain
	if len(chain) == 0 && module.MetaGen != nil {
		chain = append(chain, module.MetaGen)
	}
	if len(chain) == 0 {
		return metagen.Metadata{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]metagen.Metadata, len(chain))
	errs := make([]error, len(chain))

	var wg sync.WaitGroup
	for idx, fn := range chain {
		wg.Add(1)
		go func(i int, run PageMetaGen[C, P]) {
			defer wg.Done()
			if run == nil {
				return
			}
			startedAt := time.Now()
			meta, err := run(ctx, appCtx, r, params)
			runtime.LogResolverTiming(ResolverTiming{
				RoutePattern: module.Pattern,
				Stage:        ResolverStageMetaGen,
				Method:       metaGenMethodName(module, i, run),
				Duration:     time.Since(startedAt),
				Err:          err,
			})
			if err != nil {
				errs[i] = err
				cancel()
				return
			}
			results[i] = meta
		}(idx, fn)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return metagen.Metadata{}, err
		}
	}
	return metagen.MergeAll(results...), nil
}

func resolveMetadataWithContext[C interface{}, P interface{}](
	runtime RuntimeContext[C],
	ctx context.Context,
	appCtx C,
	r *http.Request,
	params P,
	routePattern string,
	chain []PageMetaGenContext[C, P],
	chainNames []string,
) (metagen.Metadata, error) {
	if len(chain) == 0 {
		return metagen.Metadata{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	metaCtx := metadataContextForResolve(runtime, ctx, appCtx, r)
	results := make([]metagen.Metadata, len(chain))
	errs := make([]error, len(chain))

	var wg sync.WaitGroup
	for idx, fn := range chain {
		wg.Add(1)
		go func(i int, run PageMetaGenContext[C, P]) {
			defer wg.Done()
			if run == nil {
				return
			}
			startedAt := time.Now()
			meta, err := run(metaCtx, params)
			runtime.LogResolverTiming(ResolverTiming{
				RoutePattern: routePattern,
				Stage:        ResolverStageMetaGen,
				Method:       metaGenChainMethodName(chainNames, i),
				Duration:     time.Since(startedAt),
				Err:          err,
			})
			if err != nil {
				errs[i] = err
				cancel()
				return
			}
			results[i] = meta
		}(idx, fn)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return metagen.Metadata{}, err
		}
	}
	return metagen.MergeAll(results...), nil
}

func metadataContextForResolve[C any](
	runtime RuntimeContext[C],
	ctx context.Context,
	appCtx C,
	r *http.Request,
) MetaContext[C] {
	return NewMetaContext(ctx, appCtx, r, runtime.ResolveRoot(r), runtime.I18n())
}

func handleModuleError[C interface{}](
	runtime RuntimeContext[C],
	w http.ResponseWriter,
	r *http.Request,
	err error,
	routeID string,
	routePattern string,
	source NotFoundSource,
	stage string,
) {
	if IsNotFound(err) {
		runtime.RespondNotFound(w, r, NotFoundContext{
			RequestPath:         r.URL.Path,
			MatchedRouteID:      routeID,
			MatchedRoutePattern: routePattern,
			Locale:              frameworki18n.LocaleFromContext(r.Context()),
			Source:              source,
		})
		return
	}

	runtime.RespondServerError(w, fmt.Errorf("%s route %q: %w", stage, routePattern, err))
}

func loadMethodName[C interface{}, P interface{}, VM interface{}](module PageModule[C, P, VM]) string {
	if name := strings.TrimSpace(module.LoadName); name != "" {
		return name
	}
	return resolverFuncName(module.Load)
}

func metaGenMethodName[C interface{}, P interface{}, VM interface{}](
	module PageModule[C, P, VM],
	index int,
	run PageMetaGen[C, P],
) string {
	if name := metaGenChainMethodName(module.MetaGenChainNames, index); name != "" {
		return name
	}
	if len(module.MetaGenChain) == 0 {
		if name := strings.TrimSpace(module.MetaGenName); name != "" {
			return name
		}
	}
	return resolverFuncName(run)
}

func metaGenChainMethodName(chainNames []string, index int) string {
	if index >= 0 && index < len(chainNames) {
		if name := strings.TrimSpace(chainNames[index]); name != "" {
			return name
		}
	}
	return ""
}

func resolverFuncName(fn interface{}) string {
	if fn == nil {
		return "unknown"
	}
	value := reflect.ValueOf(fn)
	if !value.IsValid() || value.Kind() != reflect.Func {
		return "unknown"
	}
	ptr := value.Pointer()
	if ptr == 0 {
		return "unknown"
	}
	resolved := goruntime.FuncForPC(ptr)
	if resolved == nil {
		return "unknown"
	}
	return resolved.Name()
}
