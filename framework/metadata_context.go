package framework

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	frameworksite "github.com/RevoTale/no-js/framework/site"
)

type BaseMetadataContext interface {
	Context() context.Context
	Request() *http.Request
	Locale() string
	Root() *url.URL
	CurrentURL() *url.URL
	URL(path string) *url.URL
	LocalizedURL(locale string, path string) *url.URL
	Alternates(locale string, types map[string]string) (metagen.Alternates, error)
}

type MetaContext[C any] interface {
	BaseMetadataContext
	App() C
}

// MetadataContext remains as a compatibility bridge while callers move to MetaContext[C].
type MetadataContext[K ~string] interface {
	BaseMetadataContext
}

type metadataContext[C any] struct {
	ctx        context.Context
	app        C
	request    *http.Request
	root       *url.URL
	i18nConfig frameworki18n.Config
	locale     string
}

type metadataRootProvider interface {
	ResolveRoot(*http.Request) *url.URL
}

type metadataSiteResolverProvider interface {
	SiteResolver() frameworksite.Resolver
}

type metadataI18nProvider interface {
	I18nConfig() frameworki18n.Config
}

func NewMetaContext[C any](ctx context.Context, appCtx C, r *http.Request) MetaContext[C] {
	if ctx == nil {
		ctx = requestContext(r)
	}

	cfg := frameworki18n.Config{}
	if provider, ok := any(appCtx).(metadataI18nProvider); ok {
		cfg = provider.I18nConfig()
	}

	return metadataContext[C]{
		ctx:        ctx,
		app:        appCtx,
		request:    r,
		root:       metadataRootForAppContext(appCtx, r),
		i18nConfig: cfg,
		locale:     metadataLocaleForRequest(r, cfg),
	}
}

// NewMetadataContext stays available for older call sites and tests that only need the URL helpers.
func NewMetadataContext[K ~string, C any](appCtx C, r *http.Request) MetadataContext[K] {
	return NewMetaContext(requestContext(r), appCtx, r)
}

func (ctx metadataContext[C]) Context() context.Context {
	return ctx.ctx
}

func (ctx metadataContext[C]) App() C {
	return ctx.app
}

func (ctx metadataContext[C]) Request() *http.Request {
	return ctx.request
}

func (ctx metadataContext[C]) Locale() string {
	return strings.TrimSpace(ctx.locale)
}

func (ctx metadataContext[C]) Root() *url.URL {
	return cloneURL(ctx.root)
}

func (ctx metadataContext[C]) CurrentURL() *url.URL {
	if ctx.request == nil || ctx.request.URL == nil {
		return nil
	}
	return joinRootAndPath(
		ctx.root,
		strings.TrimSpace(ctx.request.URL.Path),
		ctx.request.URL.Query(),
	)
}

func (ctx metadataContext[C]) URL(pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, pathValue, nil)
}

func (ctx metadataContext[C]) LocalizedURL(locale string, pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, metadataLocalizedPath(ctx.i18nConfig, locale, pathValue), nil)
}

func (ctx metadataContext[C]) Alternates(locale string, types map[string]string) (metagen.Alternates, error) {
	root := ctx.Root()
	if root == nil {
		return metagen.Alternates{}, fmt.Errorf("metadata root URL is required")
	}

	return metagen.BuildAlternates(
		root.String(),
		ctx.i18nConfig,
		locale,
		requestPathWithQuery(ctx.request),
		types,
	)
}

func metadataRootForAppContext[C any](appCtx C, r *http.Request) *url.URL {
	if provider, ok := any(appCtx).(metadataRootProvider); ok {
		return cloneURL(provider.ResolveRoot(r))
	}
	if provider, ok := any(appCtx).(metadataSiteResolverProvider); ok {
		return frameworksite.ResolveRoot(provider.SiteResolver(), r)
	}
	return nil
}

func metadataLocaleForRequest(r *http.Request, cfg frameworki18n.Config) string {
	requestLocale := ""
	if r != nil {
		requestLocale = frameworki18n.LocaleFromContext(r.Context())
	}

	normalized := strings.TrimSpace(strings.ToLower(requestLocale))
	normalizedCfg, err := frameworki18n.NormalizeConfig(cfg)
	if err != nil {
		return normalized
	}
	if normalized == "" || !slices.Contains(normalizedCfg.Locales, normalized) {
		return normalizedCfg.DefaultLocale
	}
	return normalized
}

func metadataLocalizedPath(cfg frameworki18n.Config, locale string, pathValue string) string {
	normalizedCfg, err := frameworki18n.NormalizeConfig(cfg)
	if err != nil {
		return frameworki18n.NormalizePath(pathValue)
	}
	return frameworki18n.LocalizePath(normalizedCfg, strings.TrimSpace(strings.ToLower(locale)), pathValue)
}

func requestPathWithQuery(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "/"
	}
	pathValue := strings.TrimSpace(r.URL.Path)
	if pathValue == "" {
		pathValue = "/"
	}
	if strings.TrimSpace(r.URL.RawQuery) == "" {
		return pathValue
	}
	return pathValue + "?" + strings.TrimSpace(r.URL.RawQuery)
}

func joinRootAndPath(root *url.URL, pathValue string, query url.Values) *url.URL {
	if root == nil {
		return nil
	}

	trimmedPath := frameworki18n.NormalizePath(pathValue)
	clone := *root
	base := strings.TrimSuffix(strings.TrimSpace(clone.Path), "/")

	if trimmedPath == "/" {
		if base == "" {
			clone.Path = "/"
		} else {
			clone.Path = base
		}
	} else {
		joined := path.Join(base, strings.TrimPrefix(trimmedPath, "/"))
		if !strings.HasPrefix(joined, "/") {
			joined = "/" + joined
		}
		clone.Path = joined
	}

	if query == nil {
		clone.RawQuery = ""
	} else {
		clone.RawQuery = query.Encode()
	}
	clone.Fragment = ""
	return &clone
}

func cloneURL(value *url.URL) *url.URL {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func requestContext(r *http.Request) context.Context {
	if r == nil || r.Context() == nil {
		return context.Background()
	}
	return r.Context()
}
