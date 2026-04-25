# E2E Fixture Apps

`e2e/testdata` contains real consuming apps. Each fixture is a runnable example app with its own `go.mod`, real `tool` directives, and a production-style `cmd/server` entrypoint.

The `e2e` package treats them as black-box apps:
- copy the fixture into a temp dir
- rewrite only the local `replace github.com/RevoTale/no-js => ...`
- run the real public generation flow for that app
- build `./cmd/server`
- start the compiled binary as a child process
- wait for `LISTEN_URL=http://127.0.0.1:NNNN` on stdout
- assert over real loopback HTTP requests

## Fixture App Contract

Every fixture app should:
- be generatable in place with `go tool no-js ...`
- expose `cmd/server/main.go`
- accept `-addr`
- print exactly one readiness line to stdout: `LISTEN_URL=http://127.0.0.1:NNNN`
- keep operational logs on stderr
- shut down cleanly on `SIGINT` or `SIGTERM`
- avoid `cmd/probe` or other e2e-only runtime helpers

Manual verification looks like this from a fixture root:

```bash
go tool no-js gen -root . -templ-css
go run ./cmd/server -addr 127.0.0.1:8080
```

Some fixtures still use the split generation flow in the test harness where the framework currently requires it:

```bash
go tool no-js gen routes -root .
go tool templgen -base . -path web/generated -path web/components -path web/view
go tool no-js gen assets -root . -templ-css
```

## Current Fixtures

- `routepagecssapp`
  - Route-local templ `css` declared directly in `web/routes/page.templ`
  - `no-js gen -templ-css` as the only generation command for the fixture setup
  - No `web/assets` source directory; the hashed global templ stylesheet must still be built and loaded
  - Route source directories must stay free of `*_templ.go` and `templ_css_exports_gen.go`

- `clientassetsapp`
  - Colocated route and component `.css` files with generated exported class constants
  - Colocated route and component TypeScript bundled into route-level module scripts
  - Route-level client assets are present only on pages that use them
  - 404 pages receive their own route-level stylesheet

- `templcssapp`
  - Root page rendering
  - Root metadata and `metagen.Head(...)`
  - Global templ CSS asset generation
  - Shared `web/components` CSS
  - Explicit templ CSS variants
  - HTMX partial head patching
  - Root not found rendering

- `namespacedtemplcssapp`
  - Grouped route namespaces via `_group__...`
  - Slot namespaces via `_slot__...`
  - Default slot rendering through the owning layout
  - templ CSS declared inside namespaced route packages
  - Shared `web/components` CSS plus explicit variants
  - Full-page, partial, not found, and stylesheet serving through `httpserver`

- `docsfeatureapp`
  - Built-in i18n generation and localized page routing
  - Request-aware site resolution in metadata and discovery payloads
  - `MetaContext` canonical URLs, hreflang alternates, and Open Graph URLs
  - Request-scoped cache deduplication through `framework.CachedCall(...)`
  - Grouped routes, slot defaults, dynamic params, nested not found, and error handling
  - Method routes via `route.go`
  - Discovery conventions: `robots.go`, `feed.go`, nested `feed.go`, and sitemap chunks
  - Fingerprinted static assets, public files, health endpoint, and templ CSS bundling

- `groupednamespaceapp`
  - Parallel grouped route branches collapsing into the same public `/discover/...` prefix
  - Nested `_group__...` layout composition under the same public segment
  - Different full-page HTML for `/discover/notes`, `/discover/guides`, and `/discover/tags`
  - Partial rendering for grouped routes without leaking branch layouts into the body
  - Bundled templ CSS checks against the actual generated class names rendered for each branch

- `templrulesapp`
  - Real child-process server coverage over loopback HTTP, including streamed output
  - Isolated self-contained components with local templ `css`
  - `templ.KV(...)` class toggles and CSS custom properties on component roots
  - Wrapper composition through `children...` and `templ.Attributes`
  - Named slot composition through `templ.Component` parameters
  - `templ.NewOnceHandle()` dependency output and public JS module loading
  - Stable Alpine and JS hooks through `data-*` and `x-ref`
  - Grouped route namespaces plus slot defaults
  - Metadata rendering through `metagen.Head(...)` and HTMX head patches
  - Registered dynamic templ CSS variants and unregistered inline fallback behavior
  - App-owned `POST /__e2e/release-stream` only for the stream boundary assertion

- `catchallapp`
  - `_catchall__...` route matching with joined params and depth-sensitive rendering
  - Root not-found handling for unmatched parent paths
  - No `web/assets` source directory; global templ CSS must still be bundled and loaded
  - Bundled templ CSS checks against the rendered catch-all page class

- `optionalcatchallapp`
  - `_optional_catchall__...` matching for both the empty and nested path variants
  - HTMX partial rendering for the empty catch-all route
  - No `web/assets` source directory; global templ CSS must still be bundled and loaded
  - Bundled templ CSS checks against the rendered optional catch-all page class

- `methodmatrixapp`
  - `route.go` method routing with dynamic params
  - `GET`, `POST`, `PATCH`, `DELETE`, `HEAD` fallback, `OPTIONS`, and `405 Allow`
  - A real page route in the same app to verify templ CSS asset generation still works

- `i18nprefixalwaysapp`
  - Built-in i18n with `PrefixAlways`
  - Canonical redirects from unprefixed to prefixed routes
  - Generated typed message helpers, switch URLs, and localized head metadata
  - No `web/assets` source directory; global templ CSS must still be bundled and loaded
  - Bundled templ CSS checks against the localized rendered component class

- `customruntimeapp`
  - `httpserver.CustomConfig` static prefix override
  - Custom public files directory, extra routes, and health endpoint override
  - Main middleware boundary checks so only app routes receive the middleware
  - Combined site CSS plus templ CSS asset checks under the custom hashed prefix

## Planned Fixtures

- Add focused fixtures only when a built-in feature still lacks isolated coverage or a regression is easier to diagnose outside the combined `docsfeatureapp` case.
