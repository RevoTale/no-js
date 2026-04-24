# no-js

`no-js` is an opinionated Go framework for server-rendered web applications.
It is built around a strict `web/*` app tree, a generated `App Bundle`, and a
convention-first runtime.

## What You Get

- file-system routing with generated resolver contracts and an `App Bundle`
- `httpserver.NewApp(...)` as the default runtime entrypoint
- metadata, i18n, discovery, static asset, public file, and HTMX partial
  support
- build-time configuration through `no-js.bundle.yaml`

## Start Here

For app developers:

- [App Docs Overview](docs/app/overview.md)
- [Getting Started](docs/app/getting-started.md)
- [App Conventions](docs/app/conventions.md)
- [CLI Reference](docs/app/reference/cli.md)
- [Bundle Config Reference](docs/app/reference/bundle-config.md)
- [HTTP Server Reference](docs/app/reference/httpserver.md)
- [Feature Guides](docs/app/features/overview.md)
- [Troubleshooting](docs/app/troubleshooting.md)

For contributors:

- [Developing `no-js`](docs/framework/developing-no-js.md)
- [AI Agents](docs/framework/ai-agents.md)

## Happy Path

In `no-js`, the "happy path" means the default way to build an app without
custom runtime wiring:

1. Keep the standard `web/*` app layout.
2. Run generation from the app root to produce `web/generated` and
   `web/resolvers/generated.go`.
3. Implement the generated resolver methods in `web/resolvers`.
4. Build your app context and pass the generated `App Bundle` into
   `httpserver.NewApp(...)`.

Minimum required route files:

- `web/routes/root.templ`
  Defines `templ RootLayout(meta metagen.Metadata, locale string, child templ.Component)`.
- one page route such as `web/routes/page.templ`
  Defines `templ Page(model view.YourPageView)`.
- `web/routes/404.templ`
  Defines `templ NotFound(model view.RootLayoutView, path string)`.

Minimum required app-owned view definitions:

- `web/view` must use the Go package name `view`
- `type Context`
- `func (c *Context) ResolveRoot(*http.Request) *url.URL`
- `type RootLayoutView`
- `func (view RootLayoutView) LayoutPageTitle() string`
- `func NewNotFoundView() RootLayoutView`

Optional feature hooks:

- built-in i18n apps may expose `func (c *Context) I18n(*http.Request) ...`
- apps may pass `httpserver.CustomConfig.ServerErrorPage` for a generic custom
  500 page
- static-asset helpers may expose `func SetStaticAssetBasePath(string)`
- templ CSS variants may expose `func TemplCSSVariants() []templ.CSSClass`

For each page route, run generation first, then implement the generated
resolver methods in `web/resolvers`, such as `MetaGenRootPage(...)` and
`ResolveRootPage(...)`.

```bash
go tool no-js gen -root .
```

```go
appContext := view.NewContext(...)

handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

Most apps should start there. Reach for `Custom Config` only when the default
runtime wiring is not enough.

See [Getting Started](docs/app/getting-started.md) for the smallest runnable
example and [App Conventions](docs/app/conventions.md) for the stricter
contract details.

Terminology used across the docs:

- `App Bundle`: generated route and runtime contract returned by
  `generated.Bundle(appContext)`
- `Custom Config`: app-owned hooks passed to `httpserver.NewApp(...)`
- `Site Resolver`: app-owned canonical root policy
- `Advanced composition`: app-owned runtime packages used only when the
  default path is not enough
