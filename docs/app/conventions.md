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
  Reusable templ packages and component-owned Client Assets.
- `web/i18n`
- `web/assets`
  Explicit global hashed files. Do not use this for normal route/component
  CSS or JavaScript.
- `web/public`
  Fixed-path public files.

## Strict Shape At A Glance

Use this as the first check for normal page and component work. The full
validator contract follows:

```text
web/routes/
  root.templ
  root.css
  page.templ | route.go
  layout.templ | default.templ | 404.templ
  page.css | page.{js,ts,tsx,mjs,mts}
  layout.css | layout.{js,ts,tsx,mjs,mts}
  404.css | 404.{js,ts,tsx,mjs,mts}

web/components/
  card/
    card.templ | card.go
    card.css
    card.{js,ts,tsx,mjs,mts}
    helpers.go        # private declarations only
```

`web/routes` is for generation inputs that produce endpoints, layouts, 404
pages, and route-owned Client Assets. `web/components` is a component package
tree: no root files, one component per directory, same-stem assets, and public
Go API only in the same-stem anchor file.

Generated and explicit CSS/JS use `assets.browser_targets` from
`no-js.bundle.yaml`. The default is `es2020`. Route/component CSS discovery is
not CSS-import based; explicit CSS under `web/assets` may bundle relative CSS
`@import` files because it is a manual hashed asset lane.

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

## Generation Input Validation

`no-js gen` validates `web/routes` and `web/components` before generation.
Those directories are framework input trees, not general app workspaces.

Allowed `web/routes` files:

```text
root.templ              # only at web/routes/root.templ
page.templ
layout.templ
default.templ           # only inside slot roots
404.templ
route.go
robots.go              # only at web/routes/robots.go
feed.go
sitemap.go
root.css
root.css_gen.go
page.{css,js,ts,tsx,mjs,mts}
layout.{css,js,ts,tsx,mjs,mts}
default.{css,js,ts,tsx,mjs,mts}       # only beside slot-root default.templ
404.{css,js,ts,tsx,mjs,mts}
page.{css,js,ts,tsx,mjs,mts}_gen.go
layout.{css,js,ts,tsx,mjs,mts}_gen.go
default.{css,js,ts,tsx,mjs,mts}_gen.go
404.{css,js,ts,tsx,mjs,mts}_gen.go
```

Route Client Assets are valid only when the matching `root.templ`, `page.templ`,
`layout.templ`, slot-root `default.templ`, or `404.templ` exists in the same
directory. `root.css` is shell CSS for generated pages. Page, layout, slot, and component CSS is folded into
the nearest non-root layout stylesheet, or into a page fallback stylesheet when
there is no non-root layout. Slot scripts are included at the consuming layout
boundary: if a layout declares a slot, every route using that layout receives
the slot-root layout/default script and every slot page script under that slot
root. Each route owner may have at most one script source extension because
`page.ts` and `page.tsx` both emit `page.js`.

Move route helper code to `web/resolvers`, `web/view`, or another app-owned
package outside `web/routes`. Move route images, fonts, downloads, and other
static files to `web/assets` or `web/public`.

`web/components` is a component package tree. Do not put files directly
under `web/components`; each component lives in a directory whose basename is
also the Go package name and file stem.

Allowed files inside `web/components/<name>/`:

```text
<name>.templ                       # public templ anchor
<name>.go                          # public Go anchor for code-only API or helpers
*.go                               # private support code only
*_test.go                          # tests, including external package tests
<name>.css
<name>.{js,ts,tsx,mjs,mts}         # choose one; all emit <name>.js
<name>_templ.go
templ_css_exports_gen.go
<name>.{css,js,ts,tsx,mjs,mts}_gen.go
```

Each component package must contain `<name>.templ` or `<name>.go`. Handwritten
`.templ` files must be the same-stem anchor file; split large markup into
another component package instead of adding `variants.templ` or `header.templ`.
Component CSS must use the component stem and is folded into the importing
route/layout stylesheet. Component scripts must also use the component stem, and
each component package may have only one script source because `.js`, `.ts`,
`.tsx`, `.mjs`, and `.mts` all emit `<name>.js`.

Only `<name>.go` may declare exported top-level Go declarations. Support files
like `helpers.go`, `classes.go`, or `formatting.go` are allowed for line-length
and readability policies, but their funcs, methods, types, vars, and consts
must stay private. Move public component API to `<name>.go`.

Examples:

```text
web/components/card/card.templ
web/components/card/card.go
web/components/card/helpers.go       # private declarations only
web/components/card/card.css
web/components/card/card.ts
web/components/card/header/header.templ
```

Images, fonts, downloads, docs, and JSON/YAML data files belong outside
`web/components`.

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
func (Resolver) MetaGenRootNotFound(
	meta framework.MetaContext[*view.Context],
	notFound framework.NotFoundContext,
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Not Found"}, nil
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
- templ CSS extraction opt-out with `assets.templ_css: false`

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
