# Feature Guides

Use these guides after [Getting Started](../getting-started.md) and
[App Conventions](../conventions.md).

Each page focuses on one app-facing feature and starts with the default path
before showing escape hatches.

## Start With The Right Page

- [Routing and Generation](routing-and-generation.md)
  Build the route tree, generated resolver contracts, and the `App Bundle`.
- [HTTP Server and Runtime](httpserver-and-runtime.md)
  Wire `generated.Bundle(appContext)` into `httpserver.NewApp(...)`.
- [Metadata and Head](metadata-and-head.md)
  Compose canonical URLs, alternates, and `<head>` output.
- [i18n](i18n.md)
  Add locale-prefixed routing and typed translations.
- [Discovery](discovery.md)
  Serve `robots.txt`, RSS, and XML sitemaps from route conventions.
- [Static Assets](static-assets.md)
  Bundle fingerprinted assets, public files, and optional templ CSS.
- [Site Resolution](site-resolution.md)
  Provide canonical site roots for metadata, feeds, and localized URLs.
- [Request Cache and Partials](request-cache-and-partials.md)
  Reuse work inside one request and support HTMX partial refreshes.

## Common Path

1. Build the strict `web/routes` tree.
2. Run `go tool no-js gen -root .`.
3. Wire `generated.Bundle(appContext)` into `httpserver.NewApp(...)`.
4. Add optional features only when your app needs them.

## Reference Docs

- [CLI Reference](../reference/cli.md)
- [Bundle Config Reference](../reference/bundle-config.md)
- [HTTP Server Reference](../reference/httpserver.md)
