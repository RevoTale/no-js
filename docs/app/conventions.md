# App Conventions

This page is the consuming-app contract for `no-js`.

If you stay on the default path, keep the layout and generated-code assumptions
stable and let generation own the wiring.

## Required Layout

```text
your-app/
  go.mod
  web/
    routes/
      root.templ
      page.templ
      404.templ
    generated/
    resolvers/
    view/
```

At least one `page.templ` or `route.go` must exist under `web/routes`.

Optional app-owned directories:

- `web/components`
- `web/i18n`
- `web/assets`
- `web/public`

## Route Files And Directories

Reserved route template names:

- `root.templ`
  Required at `web/routes/root.templ`. This is the only place that may render
  document-level tags such as `<html>`, `<head>`, and `<body>`.
- `page.templ`
  A page route.
- `layout.templ`
  A nested layout.
- `default.templ`
  A slot fallback. Only valid at a slot root.
- `404.templ`
  Required at the app root. Optional deeper in normal route trees when a section
  needs its own not-found UI.

Reserved route Go files:

- `route.go`
  Method-only route. It may not exist at the same route as `page.templ`.
- `robots.go`
  `robots.txt` convention.
- `feed.go`
  RSS convention.
- `sitemap.go`
  XML sitemap convention.

Reserved control directories:

- `_param__slug`
- `_catchall__slug`
- `_optional_catchall__slug`
- `_group__marketing`
- `_slot__analytics`

Any other `_...` route directory name is a generation error.

Important route rules:

- slots require a same-level owning `layout.templ`
- nested slots are not allowed
- `default.templ` is only valid inside slot directories
- `route.go`, `robots.go`, `feed.go`, and `sitemap.go` are not allowed inside
  slot directories
- route-local `components/` directories are rejected; put reusable templates
  under `web/components`

## Generated Output

`no-js` writes generated files to:

```text
web/generated/
web/resolvers/generated.go
```

Commit generated output with your app, but do not edit generated files manually.
Change source templates, resolvers, view models, or configuration, then
regenerate.

## View Contract Rules

Current generated contracts expect the `web/view` package to expose app-owned
model types, not one shared framework model.

- the Go package identifier must be `view`
- generated code expects `*view.Context` as the app context type
- `view.Context` must expose `ResolveRoot(*http.Request) *url.URL`
- `Page`, `Layout`, `Default`, and `NotFound` templates must use model types
  from `view.*`
- generation reads those template signatures and writes matching resolver
  methods to `web/resolvers/generated.go`
- server errors use the default plain `Internal Server Error` response unless
  you configure `httpserver.CustomConfig.ServerErrorPage`
- generated bundle wiring calls `view.SetStaticAssetBasePath` only when that
  function exists
- generated Client Asset helpers live beside their `.css`, `.js`, and `.ts`
  source files and use the same Go package as that directory
- generated templ CSS registration appends `view.TemplCSSVariants()` only
  when that function exists

Example route template signatures:

```templ
templ Page(model view.HomePageView) {}
templ Layout(model view.MarketingLayoutView, child templ.Component) {}
templ Default(model view.SidebarDefaultView) {}
templ NotFound(model view.NotFoundView, path string) {}
```

Example generated resolver shape:

```go
func (Resolver) ResolveRootNotFound(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	params RootParams,
) (view.NotFoundView, error) {
	return view.NotFoundView{Message: "Missing"}, nil
}
```

That is the framework contract today. Keep those names stable unless you are
also changing the generator.

Recommended pattern:

- keep templates focused on rendering
- keep data loading, URL policy, and reshaping logic in resolvers or view
  models

## Runtime Happy Path

The preferred runtime integration is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

Terminology used across the docs:

- `App Bundle`: generated route and runtime contract
- `Custom Config`: app-owned hooks passed to `httpserver.NewApp(...)`
- `Site Resolver`: app-owned canonical URL policy
- `Advanced composition`: app-owned runtime packages used only when the
  default path is not enough

## Build-Time Config Boundary

`no-js.bundle.yaml` is optional and build-time only.

Use it for:

- path overrides
- feature auto-detection overrides
- static asset manifest settings

Do not use it for:

- listen addresses
- API tokens
- analytics IDs
- middleware wiring
- site resolution policy
- app service wiring

Those belong in app-owned Go code.

## Related Docs

- [Getting Started](getting-started.md)
- [CLI Reference](reference/cli.md)
- [Bundle Config Reference](reference/bundle-config.md)
- [Feature Guides](features/overview.md)
