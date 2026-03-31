package httpserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/framework"
	frameworkdiscovery "github.com/RevoTale/no-js/framework/discovery"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/a-h/templ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func ungzipBody(t *testing.T, data []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer func() {
		_ = reader.Close()
	}()

	decoded, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(decoded)
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

func TestHTTPServerCachePoliciesAndHTMX(t *testing.T) {
	t.Parallel()

	staticDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(staticDir, "file.txt"), []byte("asset"), 0o644))

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
						return "page", nil
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
		Static: StaticMount{
			URLPrefix: "/_assets/",
			Dir:       staticDir,
		},
		CachePolicies: CachePolicies{
			HTML:           "html-cache",
			Live:           "live-cache",
			LiveNavigation: "live-nav-cache",
			Static:         "static-cache",
			Health:         "health-cache",
			Error:          "error-cache",
		},
		NotFoundPage: func(framework.NotFoundContext) templ.Component {
			return textComponent("not-found")
		},
	})
	require.NoError(t, err)

	recPage := httptest.NewRecorder()
	handler.ServeHTTP(recPage, httptest.NewRequest(http.MethodGet, "/notes", nil))
	require.Equal(t, http.StatusOK, recPage.Code)
	require.Equal(t, "html-cache", recPage.Header().Get("Cache-Control"))
	require.Contains(t, recPage.Header().Get("Vary"), "HX-Request")
	require.Equal(t, "[layout]page[/layout]", strings.TrimSpace(recPage.Body.String()))

	reqHTMX := httptest.NewRequest(http.MethodGet, "/notes", nil)
	reqHTMX.Header.Set("HX-Request", "true")
	recHTMX := httptest.NewRecorder()
	handler.ServeHTTP(recHTMX, reqHTMX)
	require.Equal(t, http.StatusOK, recHTMX.Code)
	require.Equal(t, "live-cache", recHTMX.Header().Get("Cache-Control"))
	require.Equal(t, "page", strings.TrimSpace(recHTMX.Body.String()))
	require.NotEmpty(t, strings.TrimSpace(recHTMX.Header().Get("HX-Trigger-After-Settle")))

	reqHTMXNav := httptest.NewRequest(http.MethodGet, "/notes?__live=navigation", nil)
	reqHTMXNav.Header.Set("HX-Request", "true")
	recHTMXNav := httptest.NewRecorder()
	handler.ServeHTTP(recHTMXNav, reqHTMXNav)
	require.Equal(t, http.StatusOK, recHTMXNav.Code)
	require.Equal(t, "live-nav-cache", recHTMXNav.Header().Get("Cache-Control"))
	require.NotEmpty(t, strings.TrimSpace(recHTMXNav.Header().Get("HX-Trigger-After-Settle")))

	recStatic := httptest.NewRecorder()
	handler.ServeHTTP(recStatic, httptest.NewRequest(http.MethodGet, "/_assets/file.txt", nil))
	require.Equal(t, http.StatusOK, recStatic.Code)
	require.Equal(t, "static-cache", recStatic.Header().Get("Cache-Control"))

	recHealth := httptest.NewRecorder()
	handler.ServeHTTP(recHealth, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	require.Equal(t, http.StatusOK, recHealth.Code)
	require.Equal(t, "health-cache", recHealth.Header().Get("Cache-Control"))
	require.Equal(t, "ok", strings.TrimSpace(recHealth.Body.String()))
}

func TestHTTPServerGzipCompression(t *testing.T) {
	t.Parallel()

	staticDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(staticDir, "file.txt"), []byte("asset payload"), 0o644))

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
						return "page", nil
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
		Static: StaticMount{
			URLPrefix: "/_assets/",
			Dir:       staticDir,
		},
	})
	require.NoError(t, err)

	reqPage := httptest.NewRequest(http.MethodGet, "/notes", nil)
	reqPage.Header.Set("Accept-Encoding", "gzip")
	recPage := httptest.NewRecorder()
	handler.ServeHTTP(recPage, reqPage)

	require.Equal(t, http.StatusOK, recPage.Code)
	require.Equal(t, "gzip", recPage.Header().Get("Content-Encoding"))
	require.Contains(t, recPage.Header().Get("Vary"), "Accept-Encoding")
	require.Equal(t, "[layout]page[/layout]", strings.TrimSpace(ungzipBody(t, recPage.Body.Bytes())))

	reqStatic := httptest.NewRequest(http.MethodGet, "/_assets/file.txt", nil)
	reqStatic.Header.Set("Accept-Encoding", "gzip")
	recStatic := httptest.NewRecorder()
	handler.ServeHTTP(recStatic, reqStatic)

	require.Equal(t, http.StatusOK, recStatic.Code)
	require.Equal(t, "gzip", recStatic.Header().Get("Content-Encoding"))
	require.Equal(t, "asset payload", strings.TrimSpace(ungzipBody(t, recStatic.Body.Bytes())))
}

func TestHTTPServerCanDisableHealthEndpoint(t *testing.T) {
	t.Parallel()

	handler, err := New(Config[*struct{}]{
		AppContext:     &struct{}{},
		DisableHealth:  true,
		NotFoundPage:   func(framework.NotFoundContext) templ.Component { return textComponent("not-found") },
		CachePolicies:  DefaultCachePolicies(),
		LogServerError: func(error) {},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, "not-found", strings.TrimSpace(rec.Body.String()))
}

func TestHTTPServerDoesNotCompressWithoutGzipAcceptEncoding(t *testing.T) {
	t.Parallel()

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
						return "page", nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
				},
			},
		},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, strings.TrimSpace(rec.Header().Get("Content-Encoding")))
	require.Equal(t, "page", strings.TrimSpace(rec.Body.String()))
}

func TestNewAppUsesBundleAndCustomConfig(t *testing.T) {
	t.Parallel()

	staticDir := t.TempDir()
	manifestPath := filepath.Join(staticDir, "manifest.json")
	require.NoError(t, os.WriteFile(manifestPath, []byte("{\n  \"version\": 1,\n  \"hash\": \"abc123\"\n}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(staticDir, "file.txt"), []byte("asset"), 0o644))

	publicDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "favicon.txt"), []byte("icon"), 0o644))

	var staticBasePath string
	handler, err := NewApp(Config[*struct{}]{
		App: AppBundle[*struct{}]{
			Context: &struct{}{},
			Handlers: []framework.RouteHandler[*struct{}]{
				framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
					Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
						Pattern: "/notes",
						ParseParams: func(path string) (framework.EmptyParams, bool) {
							return framework.EmptyParams{}, path == "/notes"
						},
						Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
							return "page", nil
						},
						Render: func(view string) templ.Component { return textComponent(view) },
					},
				},
			},
			NotFoundPage: func(framework.NotFoundContext) templ.Component {
				return textComponent("not-found")
			},
			OnStaticAssetBasePathResolved: func(prefix string) {
				staticBasePath = prefix
			},
		},
		Custom: CustomConfig{
			StaticAssets: &StaticAssetsConfig{
				ManifestPath: manifestPath,
				URLPrefix:    "/assets/",
			},
			PublicFiles: &PublicFilesConfig{
				Dir: publicDir,
			},
		},
	})
	require.NoError(t, err)

	recPage := httptest.NewRecorder()
	handler.ServeHTTP(recPage, httptest.NewRequest(http.MethodGet, "/notes", nil))
	require.Equal(t, http.StatusOK, recPage.Code)
	require.Equal(t, "page", strings.TrimSpace(recPage.Body.String()))

	recStatic := httptest.NewRecorder()
	handler.ServeHTTP(recStatic, httptest.NewRequest(http.MethodGet, "/assets/abc123/file.txt", nil))
	require.Equal(t, http.StatusOK, recStatic.Code)
	require.Equal(t, "asset", strings.TrimSpace(recStatic.Body.String()))

	recPublic := httptest.NewRecorder()
	handler.ServeHTTP(recPublic, httptest.NewRequest(http.MethodGet, "/favicon.txt", nil))
	require.Equal(t, http.StatusOK, recPublic.Code)
	require.Equal(t, "icon", strings.TrimSpace(recPublic.Body.String()))

	require.Equal(t, "/assets/abc123/", staticBasePath)
}

func TestNewAppRejectsNilPointerAppContext(t *testing.T) {
	t.Parallel()

	handler, err := NewApp(Config[*struct{}]{
		App: AppBundle[*struct{}]{
			Context: nil,
		},
	})

	require.Nil(t, handler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "app context is required")
}

func TestHTTPServerNotFoundContextForLoadAndUnmatched(t *testing.T) {
	t.Parallel()

	errNotFound := framework.ErrNotFound
	ctxs := make([]framework.NotFoundContext, 0, 2)

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
						return "", errNotFound
					},
					Render: func(view string) templ.Component { return textComponent(view) },
				},
			},
		},
		NotFoundPage: func(notFoundContext framework.NotFoundContext) templ.Component {
			ctxs = append(ctxs, notFoundContext)
			return textComponent("missing")
		},
		CachePolicies: CachePolicies{
			Error: "error-cache",
		},
	})
	require.NoError(t, err)

	recLoadNotFound := httptest.NewRecorder()
	handler.ServeHTTP(recLoadNotFound, httptest.NewRequest(http.MethodGet, "/notes", nil))
	require.Equal(t, http.StatusNotFound, recLoadNotFound.Code)
	require.Equal(t, "error-cache", recLoadNotFound.Header().Get("Cache-Control"))

	recUnmatched := httptest.NewRecorder()
	handler.ServeHTTP(recUnmatched, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusNotFound, recUnmatched.Code)

	require.Len(t, ctxs, 2)
	require.Equal(t, framework.NotFoundSourcePageLoad, ctxs[0].Source)
	require.Equal(t, "/notes", ctxs[0].MatchedRoutePattern)
	require.Equal(t, framework.NotFoundSourceUnmatchedRoute, ctxs[1].Source)
	require.Equal(t, "/missing", ctxs[1].RequestPath)
}

func TestHTTPServerResolverDebugToggle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		enableResolverDebug bool
		expectedEvents      int
	}{
		{
			name:                "enabled",
			enableResolverDebug: true,
			expectedEvents:      2,
		},
		{
			name:                "disabled",
			enableResolverDebug: false,
			expectedEvents:      0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			events := make(chan framework.ResolverTiming, 4)

			handler, err := New(Config[*struct{}]{
				AppContext: &struct{}{},
				Handlers: []framework.RouteHandler[*struct{}]{
					framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
						Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
							Pattern: "/notes",
							ParseParams: func(path string) (framework.EmptyParams, bool) {
								return framework.EmptyParams{}, path == "/notes"
							},
							MetaGenName: "route_resolvers.Resolver.MetaGenNotesPage",
							MetaGen: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (metagen.Metadata, error) {
								return metagen.Metadata{Title: "Notes"}, nil
							},
							LoadName: "route_resolvers.Resolver.ResolveNotesPage",
							Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
								return "page", nil
							},
							Render: func(view string) templ.Component { return textComponent(view) },
						},
					},
				},
				EnableResolverDebug: tc.enableResolverDebug,
				LogResolverTiming: func(event framework.ResolverTiming) {
					events <- event
				},
			})
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes", nil))
			require.Equal(t, http.StatusOK, rec.Code)

			out := drainResolverTimingEvents(events)
			require.Len(t, out, tc.expectedEvents)
			if tc.enableResolverDebug {
				byStage := make(map[framework.ResolverStage]framework.ResolverTiming, len(out))
				for _, event := range out {
					byStage[event.Stage] = event
				}
				metaEvent, ok := byStage[framework.ResolverStageMetaGen]
				require.True(t, ok)
				assert.Equal(t, "route_resolvers.Resolver.MetaGenNotesPage", metaEvent.Method)

				loadEvent, ok := byStage[framework.ResolverStageLoad]
				require.True(t, ok)
				assert.Equal(t, "route_resolvers.Resolver.ResolveNotesPage", loadEvent.Method)
			}
		})
	}
}

func TestHTTPServerAppliesI18nToMainRoutesOnly(t *testing.T) {
	t.Parallel()

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		I18n: &frameworki18n.Config{
			Locales:       []string{"en", "de"},
			DefaultLocale: "en",
			PrefixMode:    frameworki18n.PrefixAsNeeded,
		},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(_ context.Context, _ *struct{}, r *http.Request, _ framework.EmptyParams) (string, error) {
						return frameworki18n.LocaleFromContext(r.Context()), nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
				},
			},
		},
		MountExtraRoutes: func(mux *http.ServeMux) error {
			mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
				_, ok := frameworki18n.RequestInfoFromContext(r.Context())
				if ok {
					_, _ = io.WriteString(w, "has-locale")
					return
				}
				_, _ = io.WriteString(w, "no-locale")
			})
			return nil
		},
	})
	require.NoError(t, err)

	recLocalized := httptest.NewRecorder()
	handler.ServeHTTP(recLocalized, httptest.NewRequest(http.MethodGet, "/de/notes", nil))
	require.Equal(t, http.StatusOK, recLocalized.Code)
	require.Equal(t, "de", strings.TrimSpace(recLocalized.Body.String()))

	recExtra := httptest.NewRecorder()
	handler.ServeHTTP(recExtra, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	require.Equal(t, http.StatusOK, recExtra.Code)
	require.Equal(t, "no-locale", strings.TrimSpace(recExtra.Body.String()))
}

func TestHTTPServerKeepsExtraRoutesOutsideMainMiddlewaresAndPublicFiles(t *testing.T) {
	t.Parallel()

	publicDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "robots.txt"), []byte("public"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "site.txt"), []byte("public-site"), 0o644))

	mainMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Main-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handler, err := New(Config[*struct{}]{
		AppContext: &struct{}{},
		Handlers: []framework.RouteHandler[*struct{}]{
			framework.PageOnlyRouteHandler[*struct{}, framework.EmptyParams, string]{
				Page: framework.PageModule[*struct{}, framework.EmptyParams, string]{
					Pattern: "/notes",
					ParseParams: func(path string) (framework.EmptyParams, bool) {
						return framework.EmptyParams{}, path == "/notes"
					},
					Load: func(context.Context, *struct{}, *http.Request, framework.EmptyParams) (string, error) {
						return "page", nil
					},
					Render: func(view string) templ.Component { return textComponent(view) },
				},
			},
		},
		PublicFiles: &PublicFilesConfig{Dir: publicDir},
		MainMiddlewares: []func(http.Handler) http.Handler{
			mainMiddleware,
		},
		MountExtraRoutes: func(mux *http.ServeMux) error {
			mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, "extra")
			})
			return nil
		},
	})
	require.NoError(t, err)

	recMain := httptest.NewRecorder()
	handler.ServeHTTP(recMain, httptest.NewRequest(http.MethodGet, "/notes", nil))
	require.Equal(t, http.StatusOK, recMain.Code)
	require.Equal(t, "applied", recMain.Header().Get("X-Main-Middleware"))
	require.Equal(t, "page", strings.TrimSpace(recMain.Body.String()))

	recPublic := httptest.NewRecorder()
	handler.ServeHTTP(recPublic, httptest.NewRequest(http.MethodGet, "/site.txt", nil))
	require.Equal(t, http.StatusOK, recPublic.Code)
	require.Empty(t, strings.TrimSpace(recPublic.Header().Get("X-Main-Middleware")))
	require.Equal(t, "public-site", strings.TrimSpace(recPublic.Body.String()))

	recExtra := httptest.NewRecorder()
	handler.ServeHTTP(recExtra, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	require.Equal(t, http.StatusOK, recExtra.Code)
	require.Empty(t, strings.TrimSpace(recExtra.Header().Get("X-Main-Middleware")))
	require.Equal(t, "extra", strings.TrimSpace(recExtra.Body.String()))
}

func TestNewAppDiscoveryTakesPrecedenceOverPublicFilesAndExtraRoutes(t *testing.T) {
	t.Parallel()

	publicDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(publicDir, "robots.txt"), []byte("public"), 0o644))

	handler, err := NewApp(Config[*struct{}]{
		App: AppBundle[*struct{}]{
			Context: &struct{}{},
			Discovery: &frameworkdiscovery.Bundle[*struct{}]{
				Robots: func(framework.RuntimeContext[*struct{}], *http.Request) (frameworkdiscovery.Robots, error) {
					return frameworkdiscovery.Robots{
						Rules: []frameworkdiscovery.RobotsRule{
							{UserAgent: "*", Allow: []string{"/"}},
						},
						Sitemaps: []string{"https://example.com/sitemap-index"},
					}, nil
				},
			},
		},
		Custom: CustomConfig{
			PublicFiles: &PublicFilesConfig{Dir: publicDir},
			ExtraRoutes: func(mux *http.ServeMux) error {
				mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
					_, _ = io.WriteString(w, "extra")
				})
				return nil
			},
		},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Sitemap: https://example.com/sitemap-index")
	require.NotContains(t, rec.Body.String(), "public")
	require.NotContains(t, rec.Body.String(), "extra")
}

func drainResolverTimingEvents(events <-chan framework.ResolverTiming) []framework.ResolverTiming {
	out := make([]framework.ResolverTiming, 0, cap(events))
	for {
		select {
		case event := <-events:
			out = append(out, event)
		default:
			return out
		}
	}
}
