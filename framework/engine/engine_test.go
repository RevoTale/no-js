package engine

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog/framework"
	"blog/framework/metagen"
	"github.com/a-h/templ"
)

type testAppContext struct{}

type componentFunc func(ctx context.Context, w io.Writer) error

func (f componentFunc) Render(ctx context.Context, w io.Writer) error {
	return f(ctx, w)
}

func textComponent(value string) templ.Component {
	return componentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, value)
		return err
	})
}

func wrapComponent(tag string, child templ.Component) templ.Component {
	return componentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "["+tag+"]"); err != nil {
			return err
		}
		if err := child.Render(ctx, w); err != nil {
			return err
		}
		_, err := io.WriteString(w, "[/"+tag+"]")
		return err
	})
}

func TestServeRoutePageOnly(t *testing.T) {
	var rendered string

	routeEngine, err := New(Config[*testAppContext]{
		AppContext: &testAppContext{},
		Handlers: []framework.RouteHandler[*testAppContext]{
			framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
				Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
						return "page", nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
				},
			},
		},
		RenderPage: func(_ *http.Request, _ http.ResponseWriter, component templ.Component, _ metagen.Metadata) error {
			var b bytes.Buffer
			if err := component.Render(context.Background(), &b); err != nil {
				return err
			}
			rendered = b.String()
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
		t.Fatal("expected route to match")
	}
	if rendered != "page" {
		t.Fatalf("expected page content, got %q", rendered)
	}

	if routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil)) {
		t.Fatal("did not expect missing route to match")
	}
}

func TestServeRouteSkipsLayoutsForPartialRequests(t *testing.T) {
	var rendered string

	routeEngine, err := New(Config[*testAppContext]{
		AppContext: &testAppContext{},
		Handlers: []framework.RouteHandler[*testAppContext]{
			framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
				Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
						return "body", nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
					Layouts: []framework.LayoutRenderer[string]{
						func(_ metagen.Metadata, _ string, child templ.Component) templ.Component {
							return wrapComponent("layout", child)
						},
					},
				},
			},
		},
		IsPartialRequest: func(_ *http.Request) bool { return true },
		RenderPage: func(_ *http.Request, _ http.ResponseWriter, component templ.Component, _ metagen.Metadata) error {
			var b bytes.Buffer
			if err := component.Render(context.Background(), &b); err != nil {
				return err
			}
			rendered = b.String()
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
		t.Fatal("expected route to match")
	}
	if rendered != "body" {
		t.Fatalf("expected partial body without layout, got %q", rendered)
	}
}

func TestNotFoundAndServerErrorClassification(t *testing.T) {
	errNotFound := errors.New("not found")
	errBoom := errors.New("boom")

	t.Run("not found", func(t *testing.T) {
		notFoundCalled := false
		serverErrorCalled := false
		var notFoundContext framework.NotFoundContext

		routeEngine, err := New(Config[*testAppContext]{
			AppContext: &testAppContext{},
			Handlers: []framework.RouteHandler[*testAppContext]{
				framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
					Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
						Pattern: "/notes",
						ParseParams: func(path string) (framework.EmptyParams, bool) {
							return framework.EmptyParams{}, path == "/notes"
						},
						Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
							return "", errNotFound
						},
						Render: func(view string) templ.Component { return textComponent(view) },
					},
				},
			},
			RenderPage:      func(*http.Request, http.ResponseWriter, templ.Component, metagen.Metadata) error { return nil },
			IsNotFoundError: func(err error) bool { return errors.Is(err, errNotFound) },
			HandleNotFound: func(_ http.ResponseWriter, _ *http.Request, ctx framework.NotFoundContext) {
				notFoundCalled = true
				notFoundContext = ctx
			},
			HandleServerError: func(http.ResponseWriter, error) {
				serverErrorCalled = true
			},
		})
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}

		if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
			t.Fatal("expected route to match")
		}
		if !notFoundCalled {
			t.Fatal("expected not found callback")
		}
		if notFoundContext.Source != framework.NotFoundSourcePageLoad {
			t.Fatalf("expected not-found source %q, got %q", framework.NotFoundSourcePageLoad, notFoundContext.Source)
		}
		if notFoundContext.MatchedRoutePattern != "/notes" {
			t.Fatalf("expected matched route pattern /notes, got %q", notFoundContext.MatchedRoutePattern)
		}
		if notFoundContext.RequestPath != "/notes" {
			t.Fatalf("expected request path /notes, got %q", notFoundContext.RequestPath)
		}
		if serverErrorCalled {
			t.Fatal("did not expect server error callback")
		}
	})

	t.Run("server error", func(t *testing.T) {
		notFoundCalled := false
		serverErrorCalled := false

		routeEngine, err := New(Config[*testAppContext]{
			AppContext: &testAppContext{},
			Handlers: []framework.RouteHandler[*testAppContext]{
				framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
					Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
						Pattern: "/notes",
						ParseParams: func(path string) (framework.EmptyParams, bool) {
							return framework.EmptyParams{}, path == "/notes"
						},
						Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
							return "", errBoom
						},
						Render: func(view string) templ.Component { return textComponent(view) },
					},
				},
			},
			RenderPage:      func(*http.Request, http.ResponseWriter, templ.Component, metagen.Metadata) error { return nil },
			IsNotFoundError: func(error) bool { return false },
			HandleNotFound: func(http.ResponseWriter, *http.Request, framework.NotFoundContext) {
				notFoundCalled = true
			},
			HandleServerError: func(http.ResponseWriter, error) {
				serverErrorCalled = true
			},
		})
		if err != nil {
			t.Fatalf("new engine: %v", err)
		}

		if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
			t.Fatal("expected route to match")
		}
		if notFoundCalled {
			t.Fatal("did not expect not found callback")
		}
		if !serverErrorCalled {
			t.Fatal("expected server error callback")
		}
	})
}

func TestLayoutOrder(t *testing.T) {
	var rendered string

	routeEngine, err := New(Config[*testAppContext]{
		AppContext: &testAppContext{},
		Handlers: []framework.RouteHandler[*testAppContext]{
			framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
				Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
						return "body", nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
					Layouts: []framework.LayoutRenderer[string]{
						func(_ metagen.Metadata, _ string, child templ.Component) templ.Component {
							return wrapComponent("outer", child)
						},
						func(_ metagen.Metadata, _ string, child templ.Component) templ.Component {
							return wrapComponent("inner", child)
						},
					},
				},
			},
		},
		RenderPage: func(_ *http.Request, _ http.ResponseWriter, component templ.Component, _ metagen.Metadata) error {
			var b bytes.Buffer
			if err := component.Render(context.Background(), &b); err != nil {
				return err
			}
			rendered = b.String()
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
		t.Fatal("expected route to match")
	}
	if rendered != "[outer][inner]body[/inner][/outer]" {
		t.Fatalf("unexpected render output: %q", rendered)
	}
}

func TestMetaGenRunsBeforeLoad(t *testing.T) {
	t.Parallel()

	steps := make([]string, 0, 2)

	routeEngine, err := New(Config[*testAppContext]{
		AppContext: &testAppContext{},
		Handlers: []framework.RouteHandler[*testAppContext]{
			framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
				Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					MetaGen: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (metagen.Metadata, error) {
						steps = append(steps, "meta")
						return metagen.Metadata{Title: "Notes"}, nil
					},
					Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
						steps = append(steps, "load")
						return "body", nil
					},
					Render: func(view string) templ.Component {
						return textComponent(view)
					},
				},
			},
		},
		RenderPage: func(_ *http.Request, _ http.ResponseWriter, _ templ.Component, _ metagen.Metadata) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
		t.Fatal("expected route to match")
	}
	if len(steps) != 2 || steps[0] != "meta" || steps[1] != "load" {
		t.Fatalf("expected metagen before load, got steps=%v", steps)
	}
}

func TestMetaGenErrorSkipsLoadAndRender(t *testing.T) {
	t.Parallel()

	errNotFound := errors.New("not found")
	loadCalled := false
	renderCalled := false
	notFoundCalled := false

	routeEngine, err := New(Config[*testAppContext]{
		AppContext: &testAppContext{},
		Handlers: []framework.RouteHandler[*testAppContext]{
			framework.PageOnlyRouteHandler[*testAppContext, framework.EmptyParams, string]{
				Page: framework.PageModule[*testAppContext, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					MetaGen: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (metagen.Metadata, error) {
						return metagen.Metadata{}, errNotFound
					},
					Load: func(context.Context, *testAppContext, *http.Request, framework.EmptyParams) (string, error) {
						loadCalled = true
						return "body", nil
					},
					Render: func(view string) templ.Component {
						return textComponent(view)
					},
				},
			},
		},
		RenderPage: func(_ *http.Request, _ http.ResponseWriter, _ templ.Component, _ metagen.Metadata) error {
			renderCalled = true
			return nil
		},
		IsNotFoundError: func(err error) bool {
			return errors.Is(err, errNotFound)
		},
		HandleNotFound: func(_ http.ResponseWriter, _ *http.Request, ctx framework.NotFoundContext) {
			notFoundCalled = true
			if ctx.Source != framework.NotFoundSourceMetaGen {
				t.Fatalf("expected not-found source %q, got %q", framework.NotFoundSourceMetaGen, ctx.Source)
			}
		},
	})
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	if !routeEngine.ServeRoute(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/notes", nil)) {
		t.Fatal("expected route to match")
	}
	if loadCalled {
		t.Fatal("load should not run when metagen fails")
	}
	if renderCalled {
		t.Fatal("render callback should not run when metagen fails")
	}
	if !notFoundCalled {
		t.Fatal("expected not-found callback for metagen not found")
	}
}
