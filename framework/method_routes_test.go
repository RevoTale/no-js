package framework

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	frameworkrouter "github.com/RevoTale/no-js/framework/router"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

type testMethodRuntime struct{}

func (testMethodRuntime) AppContext() struct{} { return struct{}{} }
func (testMethodRuntime) I18n() *frameworki18n.Resolver {
	return nil
}
func (testMethodRuntime) ResolveRoot(*http.Request) *url.URL  { return nil }
func (testMethodRuntime) IsPartialRequest(*http.Request) bool { return false }
func (testMethodRuntime) RenderPage(*http.Request, http.ResponseWriter, templ.Component, metagen.Metadata) error {
	return nil
}
func (testMethodRuntime) RespondNotFound(http.ResponseWriter, *http.Request, NotFoundContext) {}
func (testMethodRuntime) RespondServerError(http.ResponseWriter, error)                       {}
func (testMethodRuntime) LogServerError(error)                                                {}
func (testMethodRuntime) LogResolverTiming(ResolverTiming)                                    {}

func parseMethodParams(pattern string) ParamsParser[SlugParams] {
	return func(requestPath string) (SlugParams, bool) {
		params, ok := frameworkrouter.MatchPathPattern(pattern, requestPath)
		if !ok {
			return SlugParams{}, false
		}
		values, ok := params["slug"]
		if !ok || len(values) == 0 {
			return SlugParams{}, false
		}
		return SlugParams{Slug: values[0]}, true
	}
}

func TestMethodOnlyRouteHandlerGet(t *testing.T) {
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
			GET: func(_ RuntimeContext[struct{}], w http.ResponseWriter, _ *http.Request, params SlugParams) error {
				w.WriteHeader(http.StatusCreated)
				_, err := w.Write([]byte(params.Slug))
				return err
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/note/hello-world", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(testMethodRuntime{}, rec, req)
	require.True(t, ok)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "hello-world", rec.Body.String())
}

func TestMethodOnlyRouteHandlerMatchPath(t *testing.T) {
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
		},
	}

	require.True(t, handler.MatchPath("/note/hello-world"))
	require.False(t, handler.MatchPath("/other"))
}

func TestMethodOnlyRouteHandlerMethodNotAllowed(t *testing.T) {
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
			GET: func(_ RuntimeContext[struct{}], w http.ResponseWriter, _ *http.Request, _ SlugParams) error {
				w.WriteHeader(http.StatusNoContent)
				return nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/note/hello-world", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(testMethodRuntime{}, rec, req)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, "GET, HEAD, OPTIONS", rec.Header().Get("Allow"))
}

func TestMethodOnlyRouteHandlerOptionsFallback(t *testing.T) {
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
			GET: func(_ RuntimeContext[struct{}], w http.ResponseWriter, _ *http.Request, _ SlugParams) error {
				w.WriteHeader(http.StatusNoContent)
				return nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodOptions, "/note/hello-world", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(testMethodRuntime{}, rec, req)
	require.True(t, ok)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "GET, HEAD, OPTIONS", rec.Header().Get("Allow"))
}

func TestMethodOnlyRouteHandlerHeadFallsBackToGet(t *testing.T) {
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
			GET: func(_ RuntimeContext[struct{}], w http.ResponseWriter, _ *http.Request, _ SlugParams) error {
				w.WriteHeader(http.StatusAccepted)
				_, err := w.Write([]byte("body"))
				return err
			},
		},
	}

	req := httptest.NewRequest(http.MethodHead, "/note/hello-world", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(testMethodRuntime{}, rec, req)
	require.True(t, ok)
	require.Equal(t, http.StatusAccepted, rec.Code)
	require.Empty(t, rec.Body.String())
}

func TestMethodOnlyRouteHandlerErrorsUseServerErrorPath(t *testing.T) {
	called := false
	runtime := testMethodRuntimeWithError{
		respondServerError: func(w http.ResponseWriter, err error) {
			called = true
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}
	handler := MethodOnlyRouteHandler[struct{}, SlugParams]{
		Route: MethodRouteModule[struct{}, SlugParams]{
			RouteID:     "note/_param__slug",
			Pattern:     "/note/_param__slug",
			ParseParams: parseMethodParams("/note/_param__slug"),
			GET: func(_ RuntimeContext[struct{}], _ http.ResponseWriter, _ *http.Request, _ SlugParams) error {
				return errors.New("boom")
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/note/hello-world", nil)
	rec := httptest.NewRecorder()

	ok := handler.TryServe(runtime, rec, req)
	require.True(t, ok)
	require.True(t, called)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

type testMethodRuntimeWithError struct {
	testMethodRuntime
	respondServerError func(http.ResponseWriter, error)
}

func (runtime testMethodRuntimeWithError) RespondServerError(w http.ResponseWriter, err error) {
	runtime.respondServerError(w, err)
}
