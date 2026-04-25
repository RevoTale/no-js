# HTTP Server Reference

`httpserver.NewApp(...)` is the app-facing runtime entrypoint.

Use it with the generated `App Bundle`:

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

`httpserver.NewApp(...)` is the supported happy path. The lower-level
`httpserver.New(...)` exists for advanced composition, but most apps should not
build the runtime from raw pieces.

## Defaults

When you stay on the default path, `httpserver.NewApp(...)` does this for you:

- mounts the generated route handlers
- enables built-in i18n middleware when the `App Bundle` contains i18n config
- serves hashed static assets from `web/assets-build/manifest.json` when the
  manifest exists
- uses `/_assets/` as the base static prefix before the manifest hash is added
- serves `web/public` at fixed request paths when the directory exists
- exposes `/healthz` with body `ok`
- adds gzip compression
- uses `public, max-age=3600, s-maxage=3600` as the default cache policy for
  HTML, static assets, health, and generic error responses

## `App Bundle`

`generated.Bundle(appContext)` returns the `App Bundle` that `NewApp(...)`
expects.

The generated bundle supplies:

- route handlers
- discovery handlers
- built-in i18n config
- site-root resolution through `appContext.ResolveRoot`
- generated not-found rendering
- templ CSS registration
- static asset base-path callback through `view.SetStaticAssetBasePath`
  when your `web/view` package defines that hook

Do not hand-build the bundle unless you are intentionally doing advanced
composition.

## `Custom Config`

`Custom Config` is the main escape hatch around the default runtime.

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
	Custom: httpserver.CustomConfig{
		MainMiddlewares: []func(http.Handler) http.Handler{
			requestIDMiddleware,
		},
	},
})
```

Supported `Custom Config` fields:

- `ExtraRoutes`
  Mount extra handlers on a dedicated `http.ServeMux`.
- `MainMiddlewares`
  Wrap the main generated route handler.
- `CachePolicies`
  Override cache headers for HTML, HTMX, static, health, and error responses.
- `StaticAssets`
  Override manifest path or base URL prefix.
- `PublicFiles`
  Override public-file directory, request-path prefix, or cache policy.
- `ServerErrorPage`
  Render a generic custom 500 page. The hook receives the error, but not the
  request, so it should not be used for localized or request-scoped content.
- `LogServerErrorEvent`
  Replace the default server-error logger with request-aware event logging.
  The event includes the error, request pointer, method, path, query, host,
  remote address, user agent, and common request ID headers.
- `LogServerError`
  Replace the default server-error logger with the legacy error-only callback.
  Ignored when `LogServerErrorEvent` is set.
- `LogResolverTiming`
  Receive resolver timing events when resolver debug is enabled.
- `EnableResolverDebug`
  Turn on resolver timing logs.
- `DisableHealth`
  Disable the health endpoint entirely.
- `HealthPath`
  Override the health endpoint path.
- `HealthBody`
  Override the health endpoint body.

## Related Runtime Types

### `StaticAssetsConfig`

- `ManifestPath`
  Path to the manifest file read at runtime.
- `URLPrefix`
  Base prefix used before the manifest hash is added.

### `PublicFilesConfig`

- `Dir`
  Filesystem directory to serve.
- `RequestPathPrefix`
  Optional prefix to add before public request paths.
- `CachePolicy`
  Optional cache header override. Default: `public, max-age=0`

### `CachePolicies`

- `HTML`
  Full-page HTML responses.
- `Live`
  HTMX partial responses.
- `LiveNavigation`
  HTMX navigation responses marked with `__live=navigation`.
- `Static`
  Hashed static assets.
- `Health`
  Health endpoint responses.
- `Error`
  Not-found and error-page responses.

## When To Stay On The Default Path

Stay on the default path when you just need:

- generated routes
- built-in i18n
- hashed static assets
- public files
- discovery conventions
- standard health and gzip behavior

Reach for `Custom Config` when you need app-owned middleware, extra routes, or
runtime overrides. Reach for `Advanced composition` only when `NewApp(...)`
cannot express the shape you need.

## Related Docs

- [HTTP Server and Runtime](../features/httpserver-and-runtime.md)
- [Static Assets](../features/static-assets.md)
- [Asset Pipeline Reference](asset-pipeline.md)
- [Site Resolution](../features/site-resolution.md)
