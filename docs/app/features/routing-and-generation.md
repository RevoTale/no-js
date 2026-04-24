# Routing And Generation

Use this guide when you are adding pages, method routes, layouts, slots, or
dynamic paths.

## What This Feature Does

`no-js` turns a strict `web/routes` tree into:

- generated route handlers
- generated resolver contracts
- generated route-param types
- an `App Bundle` that you pass into `httpserver.NewApp(...)`

You edit the route tree, resolver implementations, and view models. Generation
owns the handler wiring.

## Happy Path

```text
web/routes/
  root.templ
  404.templ
  page.templ
  author/
    _param__slug/
      page.templ
  api/
    ping/
      route.go
  _group__marketing/
    dashboard/
      layout.templ
      404.templ
      page.templ
      _slot__analytics/
        default.templ
```

Run generation:

```bash
go tool no-js gen routes -root .
```

`no-js` writes:

```text
web/generated/
web/resolvers/generated.go
```

Implement the methods declared in `web/resolvers/generated.go` from handwritten
files under `web/resolvers`. Run generation before you start writing those
resolver methods, because the generated file defines the exact params and method
signatures you need to satisfy.

Then wire the bundle into the runtime:

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

## Dynamic Routes Produce Generated Param Types

A route like:

```text
web/routes/author/_param__slug/page.templ
```

generates a route-specific params type and resolver contract:

```go
func (Resolver) ResolveAuthorParamSlugPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params AuthorParamSlugParams,
) (view.AuthorPageView, error) {
	return view.AuthorPageView{
		Heading: params.Slug,
	}, nil
}
```

The generated param shape follows the route path. Change the route tree, then
regenerate.

## Route-Local 404 Pages

`web/routes/404.templ` is the root fallback. Add another `404.templ` deeper in
the route tree when a section needs its own not-found UI:

```text
web/routes/
  404.templ
  _group__support/
    help/
      layout.templ
      page.templ
      404.templ
```

When a page resolver returns `framework.ErrNotFound`, or when an unmatched URL
maps to a section with a local `404.templ`, generated routing uses the nearest
matching not-found template and wraps it with that route's layout chain.

Route groups such as `_group__support` do not appear in the URL, but they still
participate in route ownership. In the example above, `/help/missing` can render
the `help/404.templ` page inside `help/layout.templ`.

`404.templ` declares its own model type:

```templ
templ NotFound(model view.HelpNotFoundView, path string) {
	<main>{ model.Message } { path }</main>
}
```

Generation adds a matching resolver method for that route-local 404:

```go
func (Resolver) ResolveGroupSupportHelpNotFound(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	params GroupSupportHelpParams,
) (view.HelpNotFoundView, error) {
	return view.HelpNotFoundView{Message: "Missing help page"}, nil
}
```

If the app uses built-in i18n, generated not-found rendering resolves the locale
before `Resolve...NotFound(...)` runs. Use that resolver when the 404 view model
needs translations, localized URLs, or request-scoped data.

## Method Routes

Use `route.go` for method-only endpoints:

```go
func GET(
	runtimeCtx framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params ApiPingParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
```

Supported method functions:

- `GET`
- `POST`
- `PUT`
- `PATCH`
- `DELETE`
- `HEAD`
- `OPTIONS`

`route.go` and `page.templ` are mutually exclusive at the same route.

## Control Directories

Reserved control directories shape the route tree:

- `_param__slug`
  Single dynamic path segment.
- `_catchall__slug`
  Required catch-all path segments.
- `_optional_catchall__slug`
  Optional catch-all path segments.
- `_group__marketing`
  Grouping that does not appear in the public URL.
- `_slot__analytics`
  Layout slot that does not appear in the public URL.

Any other `_...` route directory name is invalid.

## Recommended Patterns

- Keep page templates view-model-shaped.
- Keep data loading and reshaping in resolvers, not in templ files.
- Put reusable templates under `web/components`, not under `web/routes`.
- Treat generated files as build output, not handwritten source.

## Related Docs

- [Getting Started](../getting-started.md)
- [App Conventions](../conventions.md)
- [CLI Reference](../reference/cli.md)
- [HTTP Server and Runtime](httpserver-and-runtime.md)
