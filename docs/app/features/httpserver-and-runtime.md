# HTTP Server And Runtime

Use this guide when you are wiring the generated app into an HTTP server or
deciding whether you need `Custom Config`.

## What This Feature Does

`httpserver.NewApp(...)` is the default runtime entrypoint. It mounts the
generated `App Bundle`, applies built-in i18n middleware when configured, serves
static assets and public files by convention, exposes the health endpoint, and
routes discovery endpoints.

## Happy Path

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
if err != nil {
	return err
}
```

This is the preferred path for normal apps. It keeps the app contract short:

- your app constructs `appContext`
- generation constructs the `App Bundle`
- `httpserver.NewApp(...)` assembles the runtime

## Runtime Defaults

- static manifest: `web/assets-build/manifest.json`
- static base prefix: `/_assets/`
- public files: `web/public` when the directory exists
- health endpoint: `/healthz`
- health body: `ok`
- gzip compression: enabled

## Use `Custom Config` For App-Owned Hooks

Add `Custom Config` only when the default path is not enough:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
	Custom: httpserver.CustomConfig{
		MainMiddlewares: []func(http.Handler) http.Handler{
			requestIDMiddleware,
		},
	},
})
```

Good fits for `Custom Config`:

- middleware
- extra routes
- cache-policy overrides
- static asset overrides
- public-file overrides
- health endpoint overrides
- resolver debug logging

## When To Use Advanced Composition

Stay on `NewApp(...)` when you can.

Use `Advanced composition` only when your runtime shape cannot be expressed with
the generated bundle plus `Custom Config`. Most apps do not need to build the
runtime from lower-level pieces.

## Related Docs

- [HTTP Server Reference](../reference/httpserver.md)
- [Static Assets](static-assets.md)
- [Request Cache and Partials](request-cache-and-partials.md)
