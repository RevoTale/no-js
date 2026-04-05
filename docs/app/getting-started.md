# Getting Started

`no-js` is designed around a strict app layout and a convention-first runtime.
The normal flow is:

1. create the `web/*` app tree
2. generate routes and assets
3. build your app context
4. pass the generated `App Bundle` into `httpserver.NewApp(...)`

## Minimal App Shape

```text
your-app/
  go.mod
  cmd/
    server/
      main.go
  web/
    generate.go
    routes/
      root.templ
      404.templ
      error.templ
      page.templ
    generated/
    resolvers/
    view/
      context.go
      view_models.go
```

## Generation Loop

Use `no-js` as a build-time generator from the consuming app root.

```bash
go run github.com/RevoTale/no-js/cmd/no-js gen -root .
```

Common split commands:

```bash
go run github.com/RevoTale/no-js/cmd/no-js gen routes -root .
go run github.com/RevoTale/no-js/cmd/no-js gen assets -root .
go run github.com/RevoTale/no-js/cmd/no-js gen check -root .
```

A typical `web/generate.go` looks like this:

```go
package web

//go:generate go run github.com/RevoTale/no-js/cmd/no-js gen routes -root ..
//go:generate go run github.com/RevoTale/no-js/cmd/templgen -base . -path components -path generated
```

## Runtime Happy Path

The generated bundle is the app contract. `httpserver.NewApp(...)` applies the
default runtime wiring for static assets, public files, health, and discovery
conventions.

```go
appContext := runtime.NewContext(...)

handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
})
if err != nil {
	return err
}
```

Add `Custom` only when the default path is not enough:

```go
handler, err := httpserver.NewApp(httpserver.Config[*runtime.Context]{
	App: generated.Bundle(appContext),
	Custom: httpserver.CustomConfig{
		MainMiddlewares: []func(http.Handler) http.Handler{
			runtime.WithCanonicalNotesRedirects,
		},
	},
})
```

## Default Conventions

The happy path assumes:

```text
Routes: web/routes
Generated code: web/generated
View contracts: web/view
Source assets: web/assets
Built assets: web/assets-build
Static URL prefix: /_assets/
Public files: web/public
```

You can override build-time paths with `no-js.bundle.yaml`, but that is a
compatibility escape hatch, not the primary model.

## What You Edit

You usually edit:

- `web/routes/*`
- `web/resolvers/*`
- `web/view/*`
- `web/components/*`
- `web/i18n/*`

You usually do not edit:

- `web/generated/*`
- `web/resolvers/generated.go`
- `web/assets-build/*`

Change the source files, then regenerate.

## Next

- [Feature Guides Overview](features/overview.md)
- [Routing and Generation](features/routing-and-generation.md)
- [HTTP Server and Runtime](features/httpserver-and-runtime.md)
- [Metadata and Head](features/metadata-and-head.md)
- [i18n](features/i18n.md)
- [Static Assets](features/static-assets.md)
