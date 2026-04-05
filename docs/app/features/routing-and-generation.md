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
    [slug]/
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

For a dynamic page under `web/routes/note/[slug]/page.templ`, the generated
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

## Related Docs

- [Getting Started](../getting-started.md)
- [App Conventions](../conventions.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
- [Metadata and Head](metadata-and-head.md)
