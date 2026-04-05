# HTTP Server and Runtime

## What This Feature Does

`httpserver.NewApp(...)` is the default runtime entrypoint. It mounts the
generated `App Bundle`, applies i18n middleware when configured, serves static
assets and public files by convention, exposes the health endpoint, and routes
discovery endpoints.

## Modules

- `framework/httpserver`

## Happy Path

The normal server setup is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
if err != nil {
	return err
}
```

Convention defaults:

- static manifest: `web/assets-build/manifest.json`
- static prefix: `/_assets/`
- public files: `web/public`
- health endpoint: `/healthz`

## Focused Example

Use `Custom Config` only when you need app-owned hooks around the default
runtime:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
	Custom: httpserver.CustomConfig{
		MainMiddlewares: []func(http.Handler) http.Handler{
			requestIDMiddleware,
		},
	},
})
if err != nil {
	return err
}
```

`Custom Config` is the right place for middleware, cache-policy overrides,
public-file overrides, or extra routes. It is not where you replace the
generated `App Bundle`.

## When To Use `Custom Config` Or Advanced Composition

Stay on the default path if convention behavior is enough.

Use `Custom Config` for middleware, cache headers, extra routes, or static/public
overrides.

Use `Advanced composition` only when you need app-owned runtime packages that do
not fit the default wiring surface.

## Related Docs

- [Getting Started](../getting-started.md)
- [Routing and Generation](routing-and-generation.md)
- [Static Assets](static-assets.md)
- [Request Cache and Partials](request-cache-and-partials.md)
