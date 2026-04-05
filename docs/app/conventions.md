# App Conventions

This is the contract a consuming app follows when using `no-js`.

## Required Layout

```text
your-app/
  go.mod
  web/
    routes/
      root.templ
      404.templ
      error.templ
    generated/
    resolvers/
    view/
```

## Route Tree Rules

- Routes live under `web/routes`.
- Dynamic segments use `[param]` directory names.
- `root.templ` is required at `web/routes/root.templ`.
- Root `404.templ` and root `error.templ` are required.
- Route templates use fixed names: `root.templ`, `layout.templ`, `page.templ`,
  `404.templ`, and `error.templ`.
- Only `root.templ` may contain document-level tags such as `<html>`, `<head>`,
  and `<body>`.
- Route-local `components/` directories are rejected by the generator.

See [Routing and Generation](features/routing-and-generation.md) for the normal
page/resolver flow.

## Generated Output

Generated files are written to:

```text
web/generated/
web/resolvers/generated.go
```

Generated code imports:

- `web/view`
- `web/resolvers`

Do not edit generated files manually. Change the source route tree or resolver
contracts, then regenerate.

See [Routing and Generation](features/routing-and-generation.md) for the public
contract behind the generated `App Bundle`.

## View Contract Rules

Current generated contracts still expect:

- the `web/view` package to use the Go package identifier `runtime`
- page view types under `runtime.*`
- layout and not-found contracts through `runtime.RootLayoutView`

This is a framework contract today, even though the directory name is `web/view`.

See [Metadata and Head](features/metadata-and-head.md) and
[Request Cache and Partials](features/request-cache-and-partials.md) for the
request-scoped runtime surface built on top of these contracts.

## Discovery Conventions

Reserved files under `web/routes` let the app provide discovery data while the
framework owns HTTP transport and serialization:

- root `robots.go`
- root or nested `sitemap.go`
- root or nested `feed.go`

Nested discovery files inherit their route directory. Examples:

- `web/routes/feed.go` serves `/feed.xml`
- `web/routes/author/[slug]/feed.go` serves `/author/:slug/feed.xml`
- `web/routes/notes/sitemap.go` serves `/notes/sitemap.xml` and nested sitemap
  index endpoints under `/notes/...`

See `framework/discovery/discovery.go` for the field-level return contracts.

See [Discovery](features/discovery.md) for the app-facing usage model.

## Runtime Happy Path

The preferred runtime integration is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
```

Terminology used across the framework:

- `App Bundle`: the generated route/runtime contract
- `Custom Config`: isolated app-owned hooks passed to `httpserver.NewApp(...)`
- `Site Resolver`: app-owned domain and canonical-URL policy
- `Advanced composition`: any app-owned package used only when the happy path is
  not enough

See [HTTP Server and Runtime](features/httpserver-and-runtime.md) and
[Site Resolution](features/site-resolution.md).

## Build Config

`no-js.bundle.yaml` is optional and build-time only.

Use it for:

- path overrides
- feature flags used during layout resolution
- static asset build settings

Do not use it for:

- listen address
- API tokens
- analytics IDs
- site/domain runtime policy
- app service wiring

Those belong in app-owned Go runtime code.

See [i18n](features/i18n.md) and [Static Assets](features/static-assets.md) for
the main build-time features currently controlled through `no-js.bundle.yaml`.
