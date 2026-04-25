# Client Assets Pipeline Notes

This note is for framework contributors and agents changing the Client Assets
implementation. App-facing usage belongs in `docs/app/features/static-assets.md`
and app-facing lookup details belong in `docs/app/reference/asset-pipeline.md`.

## Public Contract To Preserve

- Client Assets are route-static by default.
- Generated routes inject matched asset bundles through `@metagen.Head(meta)`.
- Component rendering must not register assets per call on the normal path.
- Colocated CSS generates exported class constants beside the source file.
- Colocated scripts generate exported `XScript()` helpers beside the source
  file, but normal generated routes auto-inject scripts.
- `web/assets` remains explicit hashed files, not the route/component asset
  dependency graph.

Route-static means a route bundle is produced from the matched root template,
layout chain, page or 404 template, and the colocated assets of imported
`web/components` packages reachable from those files. If a component package is
imported, its CSS and scripts are included once for that route even when the
component renders conditionally or multiple times.

This keeps runtime overhead low, makes head output deterministic, and avoids
head/body ordering problems from render-time asset registration.

## Implementation Map

- `cmd/no-js`
  Orchestrates route generation and asset generation.
- `internal/bundler/clientassets`
  Discovers Client Assets, generates source-adjacent helpers, plans route
  bundles, stages route CSS and script bundles.
- `internal/bundler/staticassets`
  Processes the staged asset tree, fingerprints files, and writes the manifest.
- `framework/metagen`
  Resolves managed asset paths and renders them through metadata/head output.
- `framework/httpserver`
  Reads the manifest at startup and serves `web/assets-build`.

## Discovery Model

Client Asset discovery scans the configured route and component roots.

For route bundles:

- page routes include `root.templ`, the matched layout chain, and `page.templ`
- not-found routes include `root.templ`, the matched layout chain, and
  `404.templ`
- a source file owns colocated assets with the same stem, such as `page.templ`
  owning `page.css` and `page.ts`
- imported component packages contribute every colocated asset in that component
  package
- component package reachability is discovered from quoted imports in `.templ`
  and `.go` files

The import scan is intentionally not full `go/packages` type analysis. It is a
fast static scan that matches the framework direction: deterministic route
bundles, cheap generation, and no request-time asset registration.

## CSS Processing

Client Asset CSS is transformed before it is staged.

The transformer:

- parses CSS with `github.com/tdewolff/parse/v2/css`
- rewrites class selectors to generated names
- preserves declaration values, custom properties, strings, and attribute values
- rewrites selector-like at-rule preludes for `@scope` and `@supports`
- preserves non-selector at-rule names such as `@layer clientassets.e2e`
- emits source-adjacent `*.css_gen.go` files in the same Go package as the CSS
  owner

Generated class names use a deterministic key based on the app-relative source
path plus the original class name. Collisions are handled by the class allocator
with a numeric suffix, so generated class names are not reused for different
source classes.

Final CSS minification and manifest fingerprinting are still handled by the
static asset pipeline after Client Assets are staged.

## Script Processing

Client Asset scripts support `.js`, `.ts`, `.mjs`, and `.mts`.

The generator writes source-adjacent helpers such as:

```go
func PageScript() templ.Component
```

Each helper renders a `type="module"` script tag with `templ.NewOnceHandle` and
`metagen.AssetURL(...)`.

Route script bundles are built with esbuild by creating a generated entry that
imports every matched script source once. The esbuild build uses browser ESM
output and minification. TypeScript is bundled by esbuild, but typechecking stays
outside `no-js`; apps should run `tsc --noEmit` when they need type safety.

## Validation Expectations

When changing this pipeline, keep coverage for:

- exported CSS constants
- anonymized selector output, including combinators, pseudo-classes,
  pseudo-elements, `@scope`, `@supports`, and `@container`
- duplicate generated class-name avoidance
- invalid CSS reporting
- script helper generation
- route-static dependency collection
- no duplicate asset injection in rendered pages
- absence of component assets on routes that do not import the component
- `web/assets` staying separate from Client Asset module bundling

Run:

```bash
task test
```
