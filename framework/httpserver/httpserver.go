package httpserver

import (
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/RevoTale/no-js/framework"
	frameworkdiscovery "github.com/RevoTale/no-js/framework/discovery"
	"github.com/RevoTale/no-js/framework/engine"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/a-h/templ"
)

const defaultCacheControlPolicy = "public, max-age=3600, s-maxage=3600"
const defaultHealthPath = "/healthz"
const defaultHealthBody = "ok"
const defaultStaticPrefix = "/_assets/"
const liveNavigationMarkerKey = "__live"
const liveNavigationMarkerValue = "navigation"

type StaticMount struct {
	URLPrefix string
	Dir       string
}

type CachePolicies struct {
	HTML           string
	Live           string
	LiveNavigation string
	Static         string
	Health         string
	Error          string
}

func DefaultCachePolicies() CachePolicies {
	return CachePolicies{
		HTML:   defaultCacheControlPolicy,
		Live:   defaultCacheControlPolicy,
		Static: defaultCacheControlPolicy,
		Health: defaultCacheControlPolicy,
		Error:  defaultCacheControlPolicy,
	}
}

type Config[C any] struct {
	App    AppBundle[C]
	Custom CustomConfig

	AppContext    C
	ExactHandlers []framework.RouteHandler[C]
	Handlers      []framework.RouteHandler[C]
	Discovery     *frameworkdiscovery.Bundle[C]

	I18n        *frameworki18n.Config
	ResolveRoot func(r *http.Request) *url.URL
	PublicFiles *PublicFilesConfig

	MountExtraRoutes func(*http.ServeMux) error
	MainMiddlewares  []func(http.Handler) http.Handler

	Static   StaticMount
	TemplCSS *TemplCSSConfig

	CachePolicies CachePolicies

	NotFoundPage        func(appCtx C, r *http.Request, notFoundContext framework.NotFoundContext) templ.Component
	LogServerError      func(err error)
	LogResolverTiming   func(event framework.ResolverTiming)
	EnableResolverDebug bool

	DisableHealth bool
	HealthPath    string
	HealthBody    string
}

type server[C any] struct {
	cachePolicies       CachePolicies
	notFoundPage        func(appCtx C, r *http.Request, notFoundContext framework.NotFoundContext) templ.Component
	appContext          C
	logServerErr        func(err error)
	logResolverTimingFn func(event framework.ResolverTiming)
	enableResolverDebug bool
	healthPath          string
	healthBody          string
	i18nResolver        *frameworki18n.Resolver

	routeEngine *engine.Engine[C]
}

func New[C any](cfg Config[C]) (http.Handler, error) {
	cachePolicies := withDefaultPolicies(cfg.CachePolicies)
	healthPath := normalizeHealthPath(cfg.HealthPath)
	if cfg.DisableHealth {
		healthPath = ""
	}
	healthBody := strings.TrimSpace(cfg.HealthBody)
	if healthBody == "" {
		healthBody = defaultHealthBody
	}

	srv := &server[C]{
		cachePolicies:       cachePolicies,
		appContext:          cfg.AppContext,
		notFoundPage:        cfg.NotFoundPage,
		logServerErr:        cfg.LogServerError,
		logResolverTimingFn: cfg.LogResolverTiming,
		enableResolverDebug: cfg.EnableResolverDebug,
		healthPath:          healthPath,
		healthBody:          healthBody,
	}

	if cfg.I18n != nil {
		resolver, err := frameworki18n.NewResolver(*cfg.I18n)
		if err != nil {
			return nil, fmt.Errorf("create i18n resolver: %w", err)
		}
		srv.i18nResolver = resolver
	}

	routeEngine, err := engine.New(engine.Config[C]{
		AppContext:        cfg.AppContext,
		Handlers:          cfg.Handlers,
		I18nResolver:      srv.i18nResolver,
		ResolveRoot:       cfg.ResolveRoot,
		IsPartialRequest:  srv.isHTMXRequest,
		RenderPage:        srv.renderPage,
		HandleNotFound:    srv.handleNotFound,
		HandleServerError: srv.handleServerError,
		LogServerError:    srv.logServerError,
		LogResolverTiming: srv.logResolverTiming,
	})
	if err != nil {
		return nil, fmt.Errorf("create route engine: %w", err)
	}
	srv.routeEngine = routeEngine

	if err := validatePathMatchers("exact", cfg.ExactHandlers); err != nil {
		return nil, err
	}
	if err := validatePathMatchers("page", cfg.Handlers); err != nil {
		return nil, err
	}

	staticPrefix := ""
	var staticHandler http.Handler
	if strings.TrimSpace(cfg.Static.Dir) != "" {
		staticPrefix = normalizeStaticPrefix(cfg.Static.URLPrefix)
		fs := http.FileServer(http.Dir(cfg.Static.Dir))
		staticHandler = withCachePolicy(cachePolicies.Static, http.StripPrefix(staticPrefix, fs))
	}

	var mainHandler http.Handler = http.HandlerFunc(srv.handleRoute)
	for _, middleware := range cfg.MainMiddlewares {
		if middleware == nil {
			continue
		}
		mainHandler = middleware(mainHandler)
	}

	if srv.i18nResolver != nil {
		bypassPrefixes := []string{}
		if strings.TrimSpace(cfg.Static.Dir) != "" {
			bypassPrefixes = append(bypassPrefixes, staticPrefix)
		}

		bypassExact := []string{}
		if strings.TrimSpace(healthPath) != "" {
			bypassExact = append(bypassExact, healthPath)
		}

		mainHandler = frameworki18n.Middleware(frameworki18n.MiddlewareConfig{
			Resolver:       srv.i18nResolver,
			BypassPrefixes: bypassPrefixes,
			BypassExact:    bypassExact,
		})(mainHandler)
	}

	var publicFilesHandler *publicFilesHandler
	if cfg.PublicFiles != nil {
		publicFilesHandler, err = newPublicFilesHandler(*cfg.PublicFiles)
		if err != nil {
			return nil, fmt.Errorf("build public files handler: %w", err)
		}
	}

	var extraRoutes *http.ServeMux
	if cfg.MountExtraRoutes != nil {
		extraRoutes = http.NewServeMux()
		if err := cfg.MountExtraRoutes(extraRoutes); err != nil {
			return nil, fmt.Errorf("mount extra routes: %w", err)
		}
	}

	var finalHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if srv.tryServeHealth(w, r) {
			return
		}
		if tryServeStatic(staticPrefix, staticHandler, w, r) {
			return
		}
		if tryServeRouteHandlers(srv.routeEngine, cfg.ExactHandlers, w, r) {
			return
		}
		if srv.shouldServeMainRoute(cfg.Handlers, r) {
			mainHandler.ServeHTTP(w, r)
			return
		}
		if frameworkdiscovery.MaybeServeSitemapChunk(srv.routeEngine, cfg.Discovery, w, r) {
			return
		}
		if tryServeMux(extraRoutes, w, r) {
			return
		}
		if publicFilesHandler != nil && publicFilesHandler.ServeHTTP(w, r) {
			return
		}

		srv.handleNotFound(w, r, framework.NotFoundContext{
			RequestPath: normalizeRequestPath(r),
			Locale:      srv.localeForRequest(r),
			Source:      framework.NotFoundSourceUnmatchedRoute,
		})
	})

	finalHandler, err = applyTemplCSS(finalHandler, cfg.Static.URLPrefix, cfg.TemplCSS)
	if err != nil {
		return nil, err
	}

	return withGzipCompression(finalHandler), nil
}

func validatePathMatchers[C any](handlerSet string, handlers []framework.RouteHandler[C]) error {
	for idx, handler := range handlers {
		if handler == nil {
			continue
		}
		if _, ok := handler.(framework.PathMatcher); ok {
			continue
		}
		return fmt.Errorf("%s handler %d does not implement framework.PathMatcher", handlerSet, idx)
	}
	return nil
}

func tryServeRouteHandlers[C any](
	runtime framework.RuntimeContext[C],
	handlers []framework.RouteHandler[C],
	w http.ResponseWriter,
	r *http.Request,
) bool {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		if handler.TryServe(runtime, w, r) {
			return true
		}
	}
	return false
}

func tryServeStatic(staticPrefix string, staticHandler http.Handler, w http.ResponseWriter, r *http.Request) bool {
	if staticHandler == nil || r == nil || r.URL == nil {
		return false
	}
	if !strings.HasPrefix(normalizeRequestPath(r), staticPrefix) {
		return false
	}
	staticHandler.ServeHTTP(w, r)
	return true
}

func tryServeMux(mux *http.ServeMux, w http.ResponseWriter, r *http.Request) bool {
	if mux == nil || r == nil {
		return false
	}
	handler, pattern := mux.Handler(r)
	if strings.TrimSpace(pattern) == "" || handler == nil {
		return false
	}
	handler.ServeHTTP(w, r)
	return true
}

func normalizeRequestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "/"
	}
	return frameworki18n.NormalizePath(r.URL.Path)
}

func matchesRoutePath[C any](handlers []framework.RouteHandler[C], pathValue string) bool {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		matcher, ok := handler.(framework.PathMatcher)
		if !ok {
			continue
		}
		if matcher.MatchPath(pathValue) {
			return true
		}
	}
	return false
}

func (s *server[C]) handleRoute(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(s.healthPath) != "" && r.URL.Path == s.healthPath {
		s.handleHealth(w)
		return
	}

	if s.routeEngine.ServeRoute(w, r) {
		return
	}

	s.handleNotFound(w, r, framework.NotFoundContext{
		RequestPath: r.URL.Path,
		Locale:      frameworki18n.LocaleFromContext(r.Context()),
		Source:      framework.NotFoundSourceUnmatchedRoute,
	})
}

func (s *server[C]) tryServeHealth(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(s.healthPath) == "" || r == nil || r.URL == nil {
		return false
	}
	if normalizeRequestPath(r) != s.healthPath {
		return false
	}
	s.handleHealth(w)
	return true
}

func (s *server[C]) shouldServeMainRoute(handlers []framework.RouteHandler[C], r *http.Request) bool {
	matchPath, ok := s.mainRouteMatchPath(r)
	if !ok {
		return false
	}
	return matchesRoutePath(handlers, matchPath)
}

func (s *server[C]) mainRouteMatchPath(r *http.Request) (string, bool) {
	requestPath := normalizeRequestPath(r)
	if s.i18nResolver == nil {
		return requestPath, true
	}

	decision := s.i18nResolver.Resolve(requestPath)
	if decision.NotFound {
		return "", false
	}

	return decision.StrippedPath, true
}

func (s *server[C]) localeForRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if locale := frameworki18n.LocaleFromContext(r.Context()); strings.TrimSpace(locale) != "" {
		return locale
	}
	if s.i18nResolver == nil {
		return ""
	}
	decision := s.i18nResolver.Resolve(normalizeRequestPath(r))
	if decision.NotFound {
		return ""
	}
	return decision.Locale
}

func (s *server[C]) renderPage(
	r *http.Request,
	w http.ResponseWriter,
	component templ.Component,
	meta metagen.Metadata,
) error {
	cachePolicy := s.cachePolicies.HTML
	if s.isHTMXRequest(r) {
		cachePolicy = s.liveCachePolicyFor(r)
		patch, err := metagen.BuildHTMXPatch(meta)
		if err != nil {
			return fmt.Errorf("build htmx metadata patch: %w", err)
		}
		if err := metagen.WriteHTMXHeaders(w, patch); err != nil {
			return fmt.Errorf("write htmx metadata patch: %w", err)
		}
	}

	return s.renderPageWithStatus(r, w, component, 0, cachePolicy)
}

func (s *server[C]) renderPageWithStatus(
	r *http.Request,
	w http.ResponseWriter,
	component templ.Component,
	statusCode int,
	cachePolicy string,
) error {
	setCachePolicy(w, cachePolicy)
	setVaryHeader(w, "HX-Request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	return component.Render(r.Context(), w)
}

func (s *server[C]) isHTMXRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(r.Header.Get("HX-Request")), "true")
}

func (s *server[C]) liveCachePolicyFor(r *http.Request) string {
	if r != nil &&
		strings.TrimSpace(r.URL.Query().Get(liveNavigationMarkerKey)) == liveNavigationMarkerValue &&
		strings.TrimSpace(s.cachePolicies.LiveNavigation) != "" {
		return s.cachePolicies.LiveNavigation
	}

	return s.cachePolicies.Live
}

func (s *server[C]) handleNotFound(
	w http.ResponseWriter,
	r *http.Request,
	notFoundContext framework.NotFoundContext,
) {
	if s.notFoundPage == nil {
		setCachePolicy(w, s.cachePolicies.Error)
		http.NotFound(w, r)
		return
	}

	component := s.notFoundPage(s.appContext, r, notFoundContext)
	if component == nil {
		setCachePolicy(w, s.cachePolicies.Error)
		http.NotFound(w, r)
		return
	}
	if err := s.renderPageWithStatus(r, w, component, http.StatusNotFound, s.cachePolicies.Error); err != nil {
		s.handleServerError(w, fmt.Errorf("render not found page: %w", err))
	}
}

func (s *server[C]) handleServerError(w http.ResponseWriter, err error) {
	setCachePolicy(w, s.cachePolicies.Error)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	s.logServerError(err)
}

func (s *server[C]) logServerError(err error) {
	if s.logServerErr != nil {
		s.logServerErr(err)
		return
	}

	log.Printf("framework server error: %v", err)
}

func (s *server[C]) logResolverTiming(event framework.ResolverTiming) {
	if !s.enableResolverDebug {
		return
	}
	if s.logResolverTimingFn != nil {
		s.logResolverTimingFn(event)
		return
	}

	outcome := "ok"
	if event.Err != nil {
		outcome = "error: " + event.Err.Error()
	}
	log.Printf(
		"framework resolver debug: route=%q stage=%s method=%q duration=%s outcome=%s",
		strings.TrimSpace(event.RoutePattern),
		event.Stage,
		strings.TrimSpace(event.Method),
		event.Duration,
		outcome,
	)
}

func (s *server[C]) handleHealth(w http.ResponseWriter) {
	setCachePolicy(w, s.cachePolicies.Health)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(s.healthBody))
}

func normalizeStaticPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return defaultStaticPrefix
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return prefix
}

func normalizeHealthPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultHealthPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func withDefaultPolicies(policies CachePolicies) CachePolicies {
	defaults := DefaultCachePolicies()
	if strings.TrimSpace(policies.HTML) == "" {
		policies.HTML = defaults.HTML
	}
	if strings.TrimSpace(policies.Live) == "" {
		policies.Live = defaults.Live
	}
	if strings.TrimSpace(policies.Static) == "" {
		policies.Static = defaults.Static
	}
	if strings.TrimSpace(policies.Health) == "" {
		policies.Health = defaults.Health
	}
	if strings.TrimSpace(policies.Error) == "" {
		policies.Error = defaults.Error
	}
	return policies
}

func setCachePolicy(w http.ResponseWriter, policy string) {
	policy = strings.TrimSpace(policy)
	if policy == "" {
		return
	}
	w.Header().Set("Cache-Control", policy)
}

func withCachePolicy(policy string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCachePolicy(w, policy)
		next.ServeHTTP(w, r)
	})
}

func setVaryHeader(w http.ResponseWriter, header string) {
	if w == nil {
		return
	}
	header = strings.TrimSpace(header)
	if header == "" {
		return
	}

	current := strings.TrimSpace(w.Header().Get("Vary"))
	if current == "" {
		w.Header().Set("Vary", header)
		return
	}

	for _, existing := range strings.Split(current, ",") {
		if strings.EqualFold(strings.TrimSpace(existing), header) {
			return
		}
	}

	parts := strings.Split(current, ",")
	parts = append(parts, header)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	parts = slices.DeleteFunc(parts, func(value string) bool {
		return value == ""
	})
	w.Header().Set("Vary", strings.Join(parts, ", "))
}

func withGzipCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			return
		}
		setVaryHeader(w, "Accept-Encoding")
		if r == nil || r.Method == http.MethodHead || !acceptsGzip(r.Header.Get("Accept-Encoding")) {
			next.ServeHTTP(w, r)
			return
		}

		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			compress:       true,
		}
		defer func() {
			_ = gzipWriter.Close()
		}()

		next.ServeHTTP(gzipWriter, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	compress    bool
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if !w.compress || !isBodyAllowedForStatus(statusCode) {
		w.compress = false
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	header := w.Header()
	if strings.TrimSpace(header.Get("Content-Encoding")) != "" {
		w.compress = false
		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	header.Del("Content-Length")
	header.Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
	w.writer = gzip.NewWriter(w.ResponseWriter)
}

func (w *gzipResponseWriter) Write(content []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if !w.compress {
		return w.ResponseWriter.Write(content)
	}
	return w.writer.Write(content)
}

func (w *gzipResponseWriter) Flush() {
	if !w.compress {
		if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
			flusher.Flush()
		}
		return
	}

	if w.writer != nil {
		_ = w.writer.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Close() error {
	if !w.compress || w.writer == nil {
		return nil
	}
	return w.writer.Close()
}

func isBodyAllowedForStatus(statusCode int) bool {
	if statusCode >= 100 && statusCode < 200 {
		return false
	}
	return statusCode != http.StatusNoContent && statusCode != http.StatusNotModified
}

func acceptsGzip(headerValue string) bool {
	for _, part := range strings.Split(headerValue, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}

		encodingToken := token
		quality := 1.0
		if semicolon := strings.Index(token, ";"); semicolon >= 0 {
			encodingToken = strings.TrimSpace(token[:semicolon])
			params := strings.Split(token[semicolon+1:], ";")
			for _, param := range params {
				param = strings.TrimSpace(param)
				if !strings.HasPrefix(strings.ToLower(param), "q=") {
					continue
				}
				value := strings.TrimSpace(param[2:])
				if parsed, err := strconv.ParseFloat(value, 64); err == nil {
					quality = parsed
				}
			}
		}

		if strings.EqualFold(encodingToken, "gzip") && quality > 0 {
			return true
		}
	}

	return false
}
