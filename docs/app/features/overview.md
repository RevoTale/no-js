# Feature Guides Overview

These pages explain `no-js` from the consuming app's side.

Each guide follows the same pattern:

- what the feature does for your app
- which framework or generator modules own it
- the normal public contract
- one focused example

Use [Getting Started](../getting-started.md) first if you are wiring a new app.
Use [App Conventions](../conventions.md) when you need the strict `web/*`
contract.

## Feature Map

- [Routing and Generation](routing-and-generation.md)
  `internal/bundler/approutegen`, `framework/contracts`
- [HTTP Server and Runtime](httpserver-and-runtime.md)
  `framework/httpserver`
- [Metadata and Head](metadata-and-head.md)
  `framework/metadata_context`, `framework/metagen`
- [i18n](i18n.md)
  `framework/i18n`, `internal/bundler/i18ngen`, `internal/projectlayout`
- [Discovery](discovery.md)
  `framework/discovery`
- [Static Assets](static-assets.md)
  `framework/staticassets`, `framework/httpserver`, `internal/bundler/staticassets`
- [Site Resolution](site-resolution.md)
  `framework/site`, `framework/metadata_context`
- [Request Cache and Partials](request-cache-and-partials.md)
  `framework` (`request_cache.go`), `framework/metagen`, `framework/httpserver`

## How To Read The Module Names

`framework/*` packages are runtime contracts.

`internal/bundler/*` packages are generation-time modules. You do not import
them in your app, but they explain where generated files and conventions come
from.

`internal/projectlayout` is the build-time module that resolves `no-js.bundle.yaml`
and feature auto-detection.

## Related Docs

- [Getting Started](../getting-started.md)
- [App Conventions](../conventions.md)
