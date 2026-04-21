# E2E Fixture Apps

`e2e/testdata` contains real consuming apps that are generated, built, and
served through the public `no-js` toolchain. Each fixture should exercise only
built-in framework features.

## Current Fixtures

- `routepagecssapp`
  - Route-local templ `css` declared directly in `web/routes/page.templ`
  - `no-js gen -templ-css` as the only generation command for the fixture setup
  - No `web/assets` source directory; the hashed global templ stylesheet must still be built and loaded
  - Route source directories must stay free of `*_templ.go` and `templ_css_exports_gen.go`

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
  - Real HTTP requests through `httptest.Server`, including streamed partial output
  - Isolated self-contained components with local templ `css`
  - `templ.KV(...)` class toggles and CSS custom properties on component roots
  - Wrapper composition through `children...` and `templ.Attributes`
  - Named slot composition through `templ.Component` parameters
  - `templ.NewOnceHandle()` dependency output and public JS module loading
  - Stable Alpine and JS hooks through `data-*` and `x-ref`
  - Grouped route namespaces plus slot defaults
  - Metadata rendering through `metagen.Head(...)` and HTMX head patches
  - Registered dynamic templ CSS variants and unregistered inline fallback behavior

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

- Add focused fixtures only when a built-in feature still lacks isolated coverage
  or a regression is easier to diagnose outside the combined `docsfeatureapp`
  case.
