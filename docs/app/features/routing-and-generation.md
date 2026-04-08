# Routing and Generation

## What This Feature Does

`no-js` turns a strict `web/routes` tree into generated route handlers, resolver
contracts, and an `App Bundle` that you pass into `httpserver.NewApp(...)`.

The route tree is the source of truth. Generated code owns wiring. Your app owns
templates, resolvers, and view contracts.

## Modules

- `internal/bundler/approutegen`
- `framework/contracts`

## Happy Path

Routes live under `web/routes`:

```text
web/routes/
  root.templ
  404.templ
  error.templ
  page.templ
  note/
    _param__slug/
      page.templ
  api/
    health/
      route.go
  _group__marketing/
    about/
      page.templ
  dashboard/
    layout.templ
    page.templ
    _slot__analytics/
      default.templ
      page.templ
```

Generation writes:

```text
web/generated/
web/resolvers/generated.go
```

Your server wiring stays short:

```go
appContext, err := runtime.NewContext(...)
if err != nil {
	return err
}

handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
if err != nil {
	return err
}
```

## Focused Example

Control directories are reserved. The generator accepts only these forms:

- `_param__slug`
- `_catchall__slug`
- `_optional_catchall__slug`
- `_group__marketing`
- `_slot__analytics`

Any other `_...` route directory is a generation error.

## Why This Syntax Exists

`no-js` does not use Next.js symbols like `[slug]`, `(group)`, or `@slot` on
disk because route directories may also contain `.go` files such as `route.go`,
`feed.go`, and `sitemap.go`.

Those symbolic names are not valid Go package path segments. The reserved
`_...__` form keeps the route tree:

- valid for Go source packages
- unambiguous for the generator
- visually distinct from normal URL segments
- strict enough to fail fast on unknown control directories

So the tradeoff is deliberate: slightly noisier directory names in exchange for
a route tree that works with both templates and Go source files.

For a dynamic page under `web/routes/note/_param__slug/page.templ`, the generated
resolver contract uses typed params:

```go
func (Resolver) LoadNotePage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params framework.SlugParams,
) (runtime.NotePageView, error) {
	return runtime.LoadNotePage(ctx, appCtx, r, params)
}

func (Resolver) MetaGenNotePage(
	meta framework.MetaContext[*runtime.Context],
	params framework.SlugParams,
) (metagen.Metadata, error) {
	return seo.MetaGenNotePage(meta, params.Slug)
}
```

The template and route path define the shape. Generated code keeps the handler
wiring in sync.

For method-only routes, use `route.go` instead of `page.templ`:

```go
func GET(
	runtime framework.RuntimeContext[*runtime.Context],
	w http.ResponseWriter,
	r *http.Request,
	params NoteParamSlugParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
```

`route.go` and `page.templ` are mutually exclusive at the same route.

## Related Docs

- [Getting Started](../getting-started.md)
- [App Conventions](../conventions.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
- [Metadata and Head](metadata-and-head.md)
