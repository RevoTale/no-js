# Getting Started

This guide shows the shortest working path from an empty Go module to a running
`no-js` app.

## Prerequisites

- a Go module
- a server entrypoint such as `cmd/server/main.go`
- willingness to keep the strict `web/*` layout

## 1. Add The Framework And Tool

Put `no-js` in your app's `go.mod` as both a dependency and a Go tool:

```go
module example.com/your-app

go 1.25.0

require (
	github.com/RevoTale/no-js vX.Y.Z
	github.com/a-h/templ v0.3.1001
)

tool github.com/RevoTale/no-js/cmd/no-js
```

Or add the entries with Go commands:

```bash
go get github.com/RevoTale/no-js@latest
go get -tool github.com/RevoTale/no-js/cmd/no-js@latest
```

The route templates and generated code use `github.com/a-h/templ`, so your app
module needs that dependency too. You can add it explicitly or let `go mod tidy`
resolve it after you add the files below.

## 2. Create The Minimal App Tree

```text
your-app/
  go.mod
  cmd/
    server/
      main.go
  web/
    routes/
      root.templ
      page.templ
      404.templ
    resolvers/
    view/
      context.go
      view_models.go
```

## Strict App Shape

Generation validates `web/routes` and `web/components`. Keep `web/routes` for
endpoint, layout, and composition inputs only. Do not put helper Go files, docs,
images, JSON/YAML data, or `error.templ` there. Route CSS and scripts are valid
only when they match a same-directory `page.templ`, `layout.templ`, or
`404.templ`.

Keep `web/components` for component packages only. Each component lives under
`web/components/<name>/` and must have `<name>.templ` or `<name>.go`. Component
CSS and scripts must use the same stem, such as `<name>.css` or `<name>.ts`; choose at most one script extension per component package.
Support `.go` files are allowed, but exported Go API belongs in `<name>.go`;
keep helpers private. Images, fonts, downloads, docs, and data files belong in
`web/assets`, `web/public`, or another app-owned package.

## 3. Add The Minimal Files

`web/view/context.go`

```go
package view

import (
	"net/http"
	"net/url"
)

type Context struct{}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) ResolveRoot(*http.Request) *url.URL {
	root, _ := url.Parse("https://example.com")
	return root
}
```

`web/view/view_models.go`

```go
package view

type RootPageView struct {
	Heading string
}

type NotFoundView struct {
	Message string
}
```

`web/routes/root.templ`

```templ
package routes

import "github.com/RevoTale/no-js/framework/metagen"

templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) {
	<html lang={ locale }>
		<head>
			@metagen.Head(meta)
		</head>
		<body>
			@child
		</body>
	</html>
}
```

`web/routes/page.templ`

```templ
package routes

import "example.com/your-app/web/view"

templ Page(model view.RootPageView) {
	<main>
		<h1>{ model.Heading }</h1>
	</main>
}
```

`web/routes/404.templ`

```templ
package routes

import "example.com/your-app/web/view"

templ NotFound(model view.NotFoundView, path string) {
	<main>{ model.Message } { path }</main>
}
```

Run `go mod tidy` once after you add the files:

```bash
go mod tidy
```

This minimal app does not define i18n, static-asset, templ-CSS variant, or
custom 500 UI hooks. templ CSS extraction is enabled by default, but generation
does not emit `styles/templ.css` unless zero-argument templ `css {}`
declarations or `web/view.TemplCSSVariants()` exist. Server errors use the
default plain `Internal Server Error` response unless you configure
`httpserver.CustomConfig.ServerErrorPage`. Add those only when you adopt the
related feature guides.

## 4. Generate The App Bundle First

Run generation from the app root:

```bash
go tool no-js gen -root .
```

This generates:

- `web/generated/*`
- `web/resolvers/generated.go`
- built-in i18n output if `web/i18n/messages` exists
- Client Asset output if colocated `.css`, `.js`, or `.ts` files exist
- explicit global assets if `web/assets` exists

Generation also creates `web/resolvers/generated.go`. That file defines the
`Resolver` type and the exact method signatures your handwritten resolver code
must implement.

Commit generated output with your app. Do not edit it by hand; change templates,
resolvers, view models, or config, then run generation again.

A minimal generated contract for the root route looks like:

```go
type RootParams struct{}

type RouteResolver interface {
	MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error)
	MetaGenRootPage(meta framework.MetaContext[*view.Context], params RootParams) (metagen.Metadata, error)
	ResolveRootPage(
		ctx context.Context,
		appCtx *view.Context,
		r *http.Request,
		params RootParams,
	) (view.RootPageView, error)
	ResolveRootNotFound(
		ctx context.Context,
		appCtx *view.Context,
		r *http.Request,
		notFound framework.NotFoundContext,
		params RootParams,
	) (view.NotFoundView, error)
}

type Resolver struct{}
```

## 5. Implement The Generated Resolver Methods

Create `web/resolvers/root.go` after generation:

```go
package resolvers

import (
	"context"
	"net/http"

	"example.com/your-app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Home"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.RootPageView, error) {
	return view.RootPageView{
		Heading: "Hello from no-js",
	}, nil
}

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

## 6. Add The Server Entrypoint

`cmd/server/main.go`

```go
package main

import (
	"log"
	"net/http"

	generated "example.com/your-app/web/generated"
	"example.com/your-app/web/view"
	"github.com/RevoTale/no-js/framework/httpserver"
)

func main() {
	appContext := view.NewContext()

	handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
		App: generated.Bundle(appContext),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", handler))
}
```

## 7. Run The Server

```bash
go run ./cmd/server
```

Open `http://127.0.0.1:8080`.

## 8. Add `templgen` Only If You Need Extra `.templ` Packages

`go tool no-js gen` handles `web/routes` and the generated route output. If your
app also has `.templ` files under `web/components`, `web/view`, or other
app-owned packages, add the companion tool:

```go
tool (
	github.com/RevoTale/no-js/cmd/no-js
	github.com/RevoTale/no-js/cmd/templgen
)
```

```bash
go get -tool github.com/RevoTale/no-js/cmd/templgen@latest
go tool templgen -base . -path web/components -path web/view -path web/generated
```

Keep `go tool no-js` as the main entrypoint. Use `templgen` only for additional
templ packages outside `web/routes`.

## Next

- [App Conventions](conventions.md)
- [Feature Guides](features/overview.md)
- [CLI Reference](reference/cli.md)
- [Bundle Config Reference](reference/bundle-config.md)
