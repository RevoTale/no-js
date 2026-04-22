# App Docs Overview

These docs are for developers or agents building an app with `no-js`.

The default mental model is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
```

`no-js` owns route generation, convention-driven runtime wiring, and the
framework-managed features around that path. Your app owns templates,
resolver implementations, view models, runtime services, and app-specific
policy.

## What `no-js` Gives You

- strict file-system routing under `web/routes`
- generated route handlers, resolver contracts, and the `App Bundle`
- convention-first runtime wiring through `httpserver.NewApp(...)`
- metadata and `<head>` composition through `metagen`
- built-in i18n with typed message helpers
- discovery conventions for `robots.txt`, RSS, and XML sitemaps
- hashed static asset serving, fixed-path public files, and optional templ CSS
  bundling
- request-scoped caching and HTMX partial support

## What Your App Owns

- `web/routes/*` templates and route-local behavior
- `web/resolvers/*` resolver implementations
- `web/view/*` view models and runtime context
- optional `web/components/*`, `web/i18n/*`, `web/assets/*`, and `web/public/*`
- app services, middleware, domain policy, and extra routes

## Recommended Patterns

- Keep templates presentational. Shape data in resolvers and view models.
- Prefer `generated.Bundle(appContext)` consumed by `httpserver.NewApp(...)`.
- Keep `no-js.bundle.yaml` build-time only. Put runtime wiring in Go code.
- Reach for `Custom Config` before `Advanced composition`.

## Read In This Order

1. [Getting Started](getting-started.md)
   Build the smallest working app.
2. [App Conventions](conventions.md)
   Learn the strict `web/*` contract and current generated-code expectations.
3. [CLI Reference](reference/cli.md)
   Use the build-time commands correctly.
4. [Bundle Config Reference](reference/bundle-config.md)
   Override paths or feature auto-detection only when you need to.
5. [Feature Guides](features/overview.md)
   Go deeper on routing, runtime, metadata, i18n, discovery, assets, site
   resolution, and HTMX behavior.
6. [Troubleshooting](troubleshooting.md)
   Fix the common generation and startup failures quickly.

## Feature Guides

- [Routing and Generation](features/routing-and-generation.md)
- [HTTP Server and Runtime](features/httpserver-and-runtime.md)
- [Metadata and Head](features/metadata-and-head.md)
- [i18n](features/i18n.md)
- [Discovery](features/discovery.md)
- [Static Assets](features/static-assets.md)
- [Site Resolution](features/site-resolution.md)
- [Request Cache and Partials](features/request-cache-and-partials.md)

## Reference

- [CLI Reference](reference/cli.md)
- [Bundle Config Reference](reference/bundle-config.md)
- [HTTP Server Reference](reference/httpserver.md)
