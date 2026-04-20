# E2E Fixture Apps

`e2e/testdata` contains real consuming apps that are generated, built, and
served through the public `no-js` toolchain. Each fixture should exercise only
built-in framework features.

## Current Fixtures

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

## Planned Fixtures

- Add focused fixtures only when a built-in feature still lacks isolated coverage
  or a regression is easier to diagnose outside the combined `docsfeatureapp`
  case.
