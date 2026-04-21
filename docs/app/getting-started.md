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

Add `no-js` as both a normal module dependency and a pinned Go tool dependency.
For the Go-side rationale, see the official docs for [tool dependencies](https://go.dev/doc/modules/managing-dependencies#tools)
and the [`tool` directive](https://go.dev/doc/modules/gomod-ref#tool).

A minimal `go.mod` looks like:

```go
module example.com/your-app

go 1.26

require github.com/RevoTale/no-js vX.Y.Z

tool github.com/RevoTale/no-js/cmd/no-js
```

Or add the entries with Go commands:

```bash
go get -tool github.com/RevoTale/no-js/cmd/no-js@latest
```

Use `go tool no-js` as the primary build-time entrypoint from the consuming app root.

```bash
go tool no-js gen -root .
```

Common split commands:

```bash
go tool no-js gen routes -root .
go tool no-js gen assets -root .
go tool no-js gen check -root .
```

To also generate a global templ stylesheet from templ `css` components, add `-templ-css`:

```bash
go tool no-js gen -root . -templ-css
```

That flag tells `no-js` to build `styles/templ.css` from the generated templ CSS registry and pass it through the normal hashed asset pipeline. The generated `App Bundle` wires that registry into `httpserver.NewApp(...)` automatically.
Without `-templ-css`, templ component CSS stays on templ's normal render path instead of becoming a hashed static asset.

If your app also keeps `.templ` files outside generated routes, compile those as a
separate templ step. A matching `go.mod` shape is:

```go
module example.com/your-app

go 1.26

require github.com/RevoTale/no-js vX.Y.Z

tool (
	github.com/RevoTale/no-js/cmd/no-js
	github.com/RevoTale/no-js/cmd/templgen
)
```

Then run:

```bash
go get -tool github.com/RevoTale/no-js/cmd/templgen@latest
go tool templgen -base . -path web/components -path web/view -path web/generated
```

If you want a convenience wrapper, keep it thin and let it call `go tool` rather
than making `go generate` the primary interface:

```go
package web

// Optional convenience wrapper. The primary entrypoint is `go tool no-js ...`.
//go:generate go tool no-js gen -root ..
//go:generate go tool templgen -base .. -path web/components -path web/view -path web/generated
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
