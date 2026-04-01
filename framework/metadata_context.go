package framework

import (
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

type MetadataContext interface {
	Request() *http.Request
	Locale() string
	Root() *url.URL
	CurrentURL() *url.URL
	URL(path string) *url.URL
	LocalizedURL(locale string, path string) *url.URL
	Alternates(locale string, types map[string]string) (metagen.Alternates, error)
}

type metadataContext struct {
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

func NewMetadataContext[C any](appCtx C, r *http.Request) MetadataContext {
	cfg := frameworki18n.Config{}
	if provider, ok := any(appCtx).(metadataI18nProvider); ok {
		cfg = provider.I18nConfig()
	}

	return metadataContext{
		request:    r,
		root:       metadataRootForAppContext(appCtx, r),
		i18nConfig: cfg,
		locale:     metadataLocaleForRequest(r, cfg),
	}
}

func (ctx metadataContext) Request() *http.Request {
	return ctx.request
}

func (ctx metadataContext) Locale() string {
	return strings.TrimSpace(ctx.locale)
}

func (ctx metadataContext) Root() *url.URL {
	return cloneURL(ctx.root)
}

func (ctx metadataContext) CurrentURL() *url.URL {
	if ctx.request == nil || ctx.request.URL == nil {
		return nil
	}
	return joinRootAndPath(
		ctx.root,
		strings.TrimSpace(ctx.request.URL.Path),
		ctx.request.URL.Query(),
	)
}

func (ctx metadataContext) URL(pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, pathValue, nil)
}

func (ctx metadataContext) LocalizedURL(locale string, pathValue string) *url.URL {
	return joinRootAndPath(ctx.root, metadataLocalizedPath(ctx.i18nConfig, locale, pathValue), nil)
}

func (ctx metadataContext) Alternates(locale string, types map[string]string) (metagen.Alternates, error) {
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
