# Asset Pipeline Reference

Use this reference when you need the app-facing rules for generated CSS,
generated scripts, explicit hashed files, and fixed public files. For task
examples, start with [Static Assets And Client Assets](../features/static-assets.md).

## Commands

| Command | What it writes |
| --- | --- |
| `go tool no-js gen routes -root .` | generated Go route code, resolver contracts, Client Asset constants, script helpers |
| `go tool no-js gen assets -root .` | route CSS bundles, route module script bundles, explicit `web/assets` files, static manifest |
| `go tool no-js gen -root .` | both steps |

`routes` is the generation step. It writes Go code that templates and the
runtime can compile.

`assets` is the asset step. It writes browser files under `web/assets-build`
and fingerprints them through the manifest.

## Source Paths

| Source | Purpose | Auto-injected |
| --- | --- | --- |
| templ `css {}` in `.templ` files | simple component-scoped CSS | global `styles/templ.css` when declarations exist unless `assets.templ_css` is false |
| `web/routes/**/<name>.css` | route, layout, or 404 CSS with the same stem as `<name>.templ` | yes, for matched routes |
| `web/components/**/*.css` | component CSS | yes, when a matched route imports the component package |
| `web/routes/**/<name>.{js,ts,mjs,mts}` | route, layout, or 404 scripts with the same stem as `<name>.templ` | yes, for matched routes |
| `web/components/**/*.{js,ts,mjs,mts}` | component scripts | yes, when a matched route imports the component package |
| `web/assets/**/*` | explicit hashed files | no |
| `web/public/**/*` | fixed request paths | no |

## Generated CSS

templ `css {}` components and colocated `.css` files serve different needs.
Prefer templ `css {}` for simple component-scoped class declarations. Use
colocated `.css` files when you need stylesheet selectors, pseudo-classes,
pseudo-elements, combinators, at-rules, or route-level CSS.

Route-owned CSS must use the same file stem as the route template that owns it:

```text
page.templ   -> page.css
layout.templ -> layout.css
404.templ    -> 404.css
```

Component CSS is package-owned. When a matched route imports a component
package, every Client Asset in that component package is included once.

Colocated `.css` files create source-adjacent Go constants:

```go
const (
	PageShellClass = "n_a1b2c3d4"
)
```

Use the constants from templ. Do not hard-code generated class names.

Generated CSS is scoped by class-name anonymization. The original class names
stay in source files; rendered HTML and built CSS use the generated names.

templ `css {}` components are extracted by default. That writes one global
`styles/templ.css` into the hashed asset output when zero-argument templ CSS
declarations or `TemplCSSVariants()` are present. Set `assets.templ_css` to
`false` in `no-js.bundle.yaml` to keep templ CSS inline.

At runtime, `httpserver.NewApp(...)` injects that stylesheet into managed head
output for every page and suppresses the registered inline templ CSS.

## Generated Scripts

Route-owned scripts follow the same-stem rule:

```text
page.templ   -> page.ts
layout.templ -> layout.ts
404.templ    -> 404.ts
```

Colocated `.js`, `.ts`, `.mjs`, and `.mts` files create exported script helpers,
for example:

```go
func PageScript() templ.Component
```

Normal generated routes do not need to call these helpers manually. Route
generation injects matched route scripts through `@metagen.Head(meta)`.

Client Asset scripts are bundled with esbuild as browser module scripts.
Relative imports and package imports are supported when esbuild can resolve
them from the app workspace.

TypeScript is bundled, not typechecked. Run `tsc --noEmit` in app validation
when you need TypeScript typechecking.

## Explicit `web/assets`

Use `web/assets` for files that need their own hashed URL and should not be
attached to a route or component.

Files under `web/assets`:

- are included in `web/assets-build`
- are fingerprinted in the manifest
- may be minified when they are `.css`, `.js`, `.mjs`, or `.cjs`
- are not auto-injected into pages
- do not get generated CSS constants or script helpers
- are not bundled as an import graph

Package imports from `node_modules` are not supported in `web/assets` CSS or
JavaScript. Browser-resolvable relative imports can work when the imported file
is also present under `web/assets`, but `no-js` does not rewrite import paths.

## Runtime Output

Generated Client Asset bundles and explicit `web/assets` files share the same
runtime output:

```text
web/assets-build/
  manifest.json
  routes/dashboard.css
  routes/dashboard.js
  embed.js
```

At runtime, `httpserver.NewApp(...)` reads the manifest and serves files under:

```text
/_assets/<hash>/
```

## Related Docs

- [Static Assets And Client Assets](../features/static-assets.md)
- [CLI Reference](cli.md)
- [Bundle Config Reference](bundle-config.md)
- [Metadata and Head](../features/metadata-and-head.md)
