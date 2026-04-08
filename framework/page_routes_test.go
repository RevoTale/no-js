package framework

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

type testPageRuntime struct {
	partial         bool
	notFoundContext NotFoundContext
	rendered        bool
}

func (runtime *testPageRuntime) AppContext() struct{} { return struct{}{} }
func (runtime *testPageRuntime) I18n() *frameworki18n.Resolver {
	return nil
}
func (runtime *testPageRuntime) ResolveRoot(*http.Request) *url.URL  { return nil }
func (runtime *testPageRuntime) IsPartialRequest(*http.Request) bool { return runtime.partial }
func (runtime *testPageRuntime) RenderPage(
	_ *http.Request,
	_ http.ResponseWriter,
	_ templ.Component,
	_ metagen.Metadata,
) error {
	runtime.rendered = true
	return nil
}
func (runtime *testPageRuntime) RespondNotFound(_ http.ResponseWriter, _ *http.Request, notFound NotFoundContext) {
	runtime.notFoundContext = notFound
}
func (runtime *testPageRuntime) RespondServerError(http.ResponseWriter, error) {}
func (runtime *testPageRuntime) LogServerError(error)                          {}
func (runtime *testPageRuntime) LogResolverTiming(ResolverTiming)              {}

func exactEmptyParams(pattern string) ParamsParser[EmptyParams] {
	return func(requestPath string) (EmptyParams, bool) {
		return EmptyParams{}, requestPath == pattern
	}
}

func TestPageOnlyRouteHandlerKeepsInternalRouteIDOnNotFound(t *testing.T) {
	runtime := &testPageRuntime{}
	handler := PageOnlyRouteHandler[struct{}, EmptyParams, struct{}]{
		Page: PageModule[struct{}, EmptyParams, struct{}]{
			RouteID:     "_group__marketing/about",
			Pattern:     "/about",
			ParseParams: exactEmptyParams("/about"),
			Load: func(context.Context, struct{}, *http.Request, EmptyParams) (struct{}, error) {
				return struct{}{}, ErrNotFound
			},
			Render: func(struct{}) templ.Component {
				return templ.ComponentFunc(func(context.Context, io.Writer) error { return nil })
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(runtime, rec, req)
	require.True(t, ok)
	require.Equal(t, "_group__marketing/about", runtime.notFoundContext.MatchedRouteID)
	require.Equal(t, "/about", runtime.notFoundContext.MatchedRoutePattern)
}

func TestPageOnlyRouteHandlerPassesPartialFlagToCompose(t *testing.T) {
	runtime := &testPageRuntime{partial: true}
	called := false
	handler := PageOnlyRouteHandler[struct{}, EmptyParams, struct{}]{
		Page: PageModule[struct{}, EmptyParams, struct{}]{
			RouteID:     "dashboard",
			Pattern:     "/dashboard",
			ParseParams: exactEmptyParams("/dashboard"),
			Load: func(context.Context, struct{}, *http.Request, EmptyParams) (struct{}, error) {
				return struct{}{}, nil
			},
			Compose: func(
				_ context.Context,
				_ RuntimeContext[struct{}],
				_ *http.Request,
				_ metagen.Metadata,
				_ struct{},
				_ EmptyParams,
				partial bool,
			) (templ.Component, error) {
				called = true
				require.True(t, partial)
				return templ.ComponentFunc(func(context.Context, io.Writer) error { return nil }), nil
			},
			Render: func(struct{}) templ.Component {
				t.Fatal("Render should not be used when Compose is present")
				return nil
			},
			RootLayout: func(metagen.Metadata, string, templ.Component) templ.Component {
				t.Fatal("RootLayout should not be used for partial requests")
				return nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(runtime, rec, req)
	require.True(t, ok)
	require.True(t, called)
	require.True(t, runtime.rendered)
}
