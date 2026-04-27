# Troubleshooting

Use the exact error text as your starting point. These are the most common
generation and startup failures surfaced by `no-js`.

## `strict routes root missing` Or `strict view root missing`

Cause:

- the configured app layout does not contain the required route tree or view
  package

Fix:

- create the missing directory
- check `project.routes_dir` and `project.view_dir` in
  `no-js.bundle.yaml`
- make sure the configured paths stay inside the app root

## `root template is required` Or `root 404 template is required`

Cause:

- the root route files are missing from `web/routes`

Fix:

- add `web/routes/root.templ`
- add `web/routes/404.templ`

These files are required on the happy path. `web/routes/error.templ` is not a
framework file and is rejected by generation; use
`httpserver.CustomConfig.ServerErrorPage` when you need custom 500 UI.

## `unsupported file in web/routes`, `unsupported route template`, Or `route.go is not allowed inside slot directories`

Cause:

- the route tree contains a file or slot layout that breaks the route contract
- `web/routes` contains helper code, docs, static files, generated templ output,
  route Client Assets without matching templates, or legacy `error.templ`

Fix:

- keep route templates limited to `root.templ`, `page.templ`, `layout.templ`,
  `default.templ`, and `404.templ`
- keep `default.templ` only at a slot root
- keep `route.go`, `robots.go`, `feed.go`, and `sitemap.go` out of slot
  directories
- do not place `page.templ` and `route.go` at the same route
- keep `page.css` with `page.templ`, `layout.css` with `layout.templ`,
  and `404.css` with `404.templ`; use the same rule for JS/TS files
- move helper code to `web/resolvers`, `web/view`, or another app package
- move static files to `web/assets` or `web/public`
- delete `web/routes/error.templ` and configure custom 500 UI through
  `httpserver.CustomConfig.ServerErrorPage`

See [App Conventions](conventions.md) for the allowed shapes.

## `unsupported file in web/components` Or `component package ... must contain`

Cause:

- files were placed directly under `web/components` instead of a component
  package directory
- a component package is missing its `<name>.templ` or `<name>.go` anchor
- a component package has extra handwritten `.templ` files such as
  `variants.templ`
- route or component CSS/scripts do not use the required same stem
- a route owner or component package contains more than one same-stem script source, such as `page.ts` and `page.tsx` beside `page.templ`
- a support `.go` file declares exported Go API
- `web/components` contains docs, images, fonts, JSON/YAML data, or other
  non-component files

Fix:

- use `web/components/<name>/<name>.templ` or
  `web/components/<name>/<name>.go` for every component package
- split large markup into more component packages, such as
  `web/components/card/header/header.templ`
- keep route and component Client Assets same-stem, such as `<name>.css` and
  `<name>.ts`; choose only one script source extension per route owner or
  component package
- move exported Go API to `<name>.go`; keep support files private
- move images, fonts, downloads, docs, and data files to `web/assets`,
  `web/public`, or another app-owned package

## `component templates must be under web/components`

Cause:

- the route tree contains a `components/` directory under `web/routes`

Fix:

- move reusable templates to `web/components`
- keep `web/routes` for route templates and reserved route Go files only

## `bundle config ... must declare version: 1`, `field unknown not found`, Or `invalid feature mode`

Cause:

- `no-js.bundle.yaml` is invalid

Fix:

- add `version: 1`
- remove unknown fields
- keep feature modes limited to `auto`, `enabled`, or `disabled`

See [Bundle Config Reference](reference/bundle-config.md) for the supported
shape.

## `strict i18n root missing`, `strict i18n messages root missing`, `invalid locale`, Or `duplicate message id`

Cause:

- built-in i18n is enabled, but the configured files or message data are
  invalid

Fix:

- create the configured i18n root and `messages/` directory
- keep locale codes to two-letter lowercase forms such as `en` or `de`
- name message files `*.en.json`, `*.de.json`, and so on
- keep message IDs unique across merged locale shards
- if you see an error about `args` outside the canonical locale, move `args`
  back to the locale file that defines the full message schema

See [i18n](features/i18n.md) for the working layout.

## `bundle client script ... Could not resolve`

Cause:

- a colocated `.js` or `.ts` Client Asset imports a package that is not
  installed in the app workspace
- the import path is only valid for browser runtime, not for esbuild bundling

Fix:

- install the package in the app workspace
- use a relative import for app-owned script modules
- keep external browser-only modules under `web/assets` and include them
  intentionally instead of importing them from a Client Asset script

See [Static Assets And Client Assets](features/static-assets.md) for the import
rules.

## CSS From `node_modules` Is Missing

Cause:

- Client Asset CSS does not use CSS `@import` as a dependency graph
- explicit `web/assets` CSS only bundles browser-resolvable relative CSS imports
- package imports such as `@import "package/styles.css"` are not resolved from
  `node_modules`

Fix:

- put route/component CSS beside the route, layout, 404 page, or component that
  owns it
- put manual global CSS under `web/assets` and include it intentionally from
  metadata or app-owned head code
- keep relative CSS imports inside `web/assets`, such as `@import "./reset.css"`
- vendor third-party CSS into `web/assets` instead of importing it from
  `node_modules`

See [Asset Pipeline Reference](reference/asset-pipeline.md) for the compile and
bundle split.

## `generated outputs differ from git state`

Cause:

- `go tool no-js gen check -root .` regenerated files and `git diff` is no
  longer clean

Fix:

- run `go tool no-js gen -root .`
- review and commit the generated changes
- rerun `go tool no-js gen check -root .`

## `public dir ... is not a directory` Or `manifest hash is required`

Cause:

- the runtime static/public-file configuration points at invalid output

Fix:

- make sure the configured public directory exists and is a directory
- make sure asset generation wrote a valid manifest file
- regenerate assets with `go tool no-js gen assets -root .`
- if you override `ManifestPath`, confirm the runtime and build-time paths match

## `metadata root URL is required`

Cause:

- metadata code called `meta.Alternates(...)` or URL helpers without a valid
  absolute site root

Fix:

- implement a valid `Site Resolver`
- make sure `appContext.ResolveRoot(*http.Request)` returns an absolute URL
- keep canonical-root policy in app-owned runtime code

See [Site Resolution](features/site-resolution.md).

## `app context is required`

Cause:

- `httpserver.NewApp(...)` received a nil app context in the `App Bundle`

Fix:

- construct `appContext` first
- pass `generated.Bundle(appContext)` into `httpserver.NewApp(...)`
- avoid calling `NewApp(...)` with a nil `*view.Context`

## `no templ files found`

Cause:

- `go tool templgen` was pointed at directories or files that do not contain
  `.templ` sources

Fix:

- check every `-path` and `-file` argument
- keep `templgen` for app-owned templ packages outside `web/routes`
- keep `go tool no-js gen` as the main build-time command

## Related Docs

- [Getting Started](getting-started.md)
- [App Conventions](conventions.md)
- [CLI Reference](reference/cli.md)
