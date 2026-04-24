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
required framework file; use `httpserver.CustomConfig.ServerErrorPage` when you
need custom 500 UI.

## `unsupported route template`, `default.templ is only allowed...`, Or `route.go is not allowed inside slot directories`

Cause:

- the route tree contains a file or slot layout that breaks the route contract

Fix:

- keep route templates limited to `root.templ`, `page.templ`, `layout.templ`,
  `default.templ`, and `404.templ`
- keep `default.templ` only at a slot root
- keep `route.go`, `robots.go`, `feed.go`, and `sitemap.go` out of slot
  directories
- do not place `page.templ` and `route.go` at the same route

See [App Conventions](conventions.md) for the allowed shapes.

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
