package framework

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
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

type metadataContext[C any] struct {
	ctx     context.Context
	app     C
	request *http.Request
	root    *url.URL
	i18n    *frameworki18n.Resolver
	locale  string
}

func NewMetaContext[C any](
	ctx context.Context,
	appCtx C,
	r *http.Request,
	root *url.URL,
	i18n *frameworki18n.Resolver,
) MetaContext[C] {
	if ctx == nil {
		ctx = requestContext(r)
	}

	return metadataContext[C]{
		ctx:     ctx,
		app:     appCtx,
		request: r,
		root:    cloneURL(root),
		i18n:    i18n,
		locale:  metadataLocaleForRequest(r, i18n),
	}
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
	return joinRootAndPath(ctx.root, metadataLocalizedPath(ctx.i18n, locale, pathValue), nil)
}

func (ctx metadataContext[C]) Alternates(locale string, types map[string]string) (metagen.Alternates, error) {
	root := ctx.Root()
	if root == nil {
		return metagen.Alternates{}, fmt.Errorf("metadata root URL is required")
	}
	if ctx.i18n == nil {
		return metagen.Alternates{}, fmt.Errorf("metadata i18n resolver is required")
	}

	return metagen.BuildAlternates(
		root.String(),
		ctx.i18n.Config(),
		locale,
		requestPathWithQuery(ctx.request),
		types,
	)
}

func metadataLocaleForRequest(r *http.Request, i18n *frameworki18n.Resolver) string {
	if i18n == nil {
		if r == nil {
			return ""
		}
		return strings.TrimSpace(strings.ToLower(frameworki18n.LocaleFromContext(r.Context())))
	}

	return i18n.LocaleForRequest(r)
}

func metadataLocalizedPath(i18n *frameworki18n.Resolver, locale string, pathValue string) string {
	if i18n == nil {
		return frameworki18n.NormalizePath(pathValue)
	}
	return i18n.LocalizedPath(strings.TrimSpace(strings.ToLower(locale)), pathValue)
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
