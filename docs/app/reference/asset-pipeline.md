# Asset Pipeline Reference

Use this reference when you need to decide where an app asset belongs, which
command generates it, and how `no-js` serves it at runtime. For task-oriented
examples, start with [Static Assets And Client Assets](../features/static-assets.md).

## Quick Choice

| You need | Put it here | Generated behavior |
| --- | --- | --- |
| Simple templ component CSS | templ `css {}` in a `.templ` file | extracted to `styles/templ.css` by default when declarations exist |
| Root CSS | `web/routes/root.css` | emitted as shell CSS and injected into generated routes when non-empty |
| Route, layout, or 404 CSS | `web/routes/**/page.css`, `layout.css`, or `404.css` | folded into root, layout-subtree, or page-fallback stylesheets |
| Component CSS | `web/components/<name>/<name>.css` | folded into the matched layout-subtree or page-fallback stylesheet when imported |
| Route, layout, or 404 JavaScript or TypeScript | `web/routes/**/page.{js,ts,tsx,mjs,mts}`, `layout.*`, or `404.*` | emitted as a shared owner module entry and injected into matched routes |
| Component JavaScript or TypeScript | one of `web/components/<name>/<name>.{js,ts,tsx,mjs,mts}` | emitted once and injected into routes that import that component package |
| Fingerprinted images, fonts, downloads, vendor files, or manual global files | `web/assets/**/*` | copied or minified into `web/assets-build` and manifest-addressed |
| Fixed paths such as `/favicon.ico` | `web/public/**/*` | served at the same request path |

Most app pages should rely on generated route/component Client Assets and
`@metagen.Head(meta)`. Use `web/assets` only when the file needs its own hashed
URL outside the route graph.

Client Assets are route-static. CSS is grouped by route shape: `root.css` is
shell CSS, non-root layouts own subtree stylesheets, and pages without a
non-root layout get a page fallback stylesheet. JavaScript stays owner-based:
matched routes inject the layout, page, 404, and imported component module
entries they need, with shared imports emitted as esbuild chunks.

## Commands

| Command | What it writes |
| --- | --- |
| `go tool no-js gen routes -root .` | generated route Go code, resolver contracts, source-adjacent Client Asset helpers, built-in i18n output, and the templ CSS registry |
| `go tool no-js gen assets -root .` | browser CSS files, script entries, shared chunks, explicit `web/assets` output, and `web/assets-build/manifest.json` |
| `go tool no-js gen -root .` | both route generation and asset generation |

`routes` writes Go files that the app compiles. `assets` writes browser files
that the runtime serves. In normal app development, run the combined command:

```bash
go tool no-js gen -root .
```

## Source Rules

### templ `css {}`

Default: enabled.

```yaml
version: 1

assets:
  templ_css: true
```

You usually omit this config because `true` is the default. When templ CSS
extraction is enabled, `no-js` scans templ `css {}` declarations and
`web/view.TemplCSSVariants()`. If zero-argument declarations or variants exist,
asset generation emits:

```text
web/assets-build/styles/templ.css
```

The stylesheet is injected into managed head output for every page, and templ's
inline CSS registration is suppressed. If no declarations or variants exist, no
`styles/templ.css` file is emitted.

Disable extraction when you intentionally want templ's inline registration path:

```yaml
version: 1

assets:
  templ_css: false
```

### Route Client Assets

Route Client Assets must be colocated with the route template that owns them.
Only these stems are valid:

```text
root.templ   -> root.css
page.templ   -> page.css   or page.{js,ts,tsx,mjs,mts}
layout.templ -> layout.css or layout.{js,ts,tsx,mjs,mts}
404.templ    -> 404.css    or 404.{js,ts,tsx,mjs,mts}
```

A route asset is valid only when the matching template exists in the same
directory. For example, `web/routes/dashboard/page.css` requires
`web/routes/dashboard/page.templ`.

Use route assets for CSS or scripts that belong to the app shell, endpoint,
layout, or 404 page. Choose only one script source extension per route owner;
for example, do not keep both `page.ts` and `page.tsx`.

CSS output follows the layout-subtree model:

- `root.css` emits `routes/root.css` and is shell CSS only
- a non-root `layout.templ` owns `routes/<layout-dir>/layout.css`, even when no
  physical `layout.css` exists
- descendant page CSS, slot CSS, and imported component CSS fold into that
  nearest non-root layout stylesheet
- if no non-root layout owns the page, page CSS and imported component CSS fold
  into a page fallback stylesheet such as `routes/dashboard/page.css`

Scripts are injected as owner module entries through `@metagen.Head(meta)`.
Slot assets are folded into the layout that can render the slot.

### Component Client Assets

Component Client Assets must live in a strict component package:

```text
web/components/meter/meter.templ
web/components/meter/meter.css
web/components/meter/meter.tsx
```

Rules:

- the component package directory name is the anchor name
- the package must contain `meter.templ` or `meter.go`
- component CSS must be `meter.css`
- component scripts must be one of `meter.js`, `meter.ts`, `meter.tsx`, `meter.mjs`, or
  `meter.mts`
- do not place files directly under `web/components`
- do not add extra handwritten templates such as `variants.templ`
- keep images, fonts, downloads, docs, and data files outside `web/components`

When a matched route imports the component package, component CSS is folded into
the nearest layout-subtree or page-fallback stylesheet. Component scripts remain
owner module entries such as `components/meter/meter.js`. Routes that do not
import the package do not receive those assets.

## Generated CSS

Colocated `.css` files create source-adjacent Go constants:

```go
const (
	PageShellClass = "n_a1b2c3d4"
)
```

Use the generated constants from templ:

```templ
templ Page() {
	<main class={ PageShellClass }>
		Dashboard
	</main>
}
```

Do not hard-code generated class names. The source CSS keeps readable class
names, while rendered HTML and built CSS use anonymized names.

Each non-empty CSS source contributes to a generated stylesheet. The output path
is chosen by route ownership, not always by the source file path:

- `web/routes/root.css` emits `routes/root.css`
- `web/routes/dashboard/layout.templ` can emit `routes/dashboard/layout.css`
  even when there is no physical `layout.css`, because descendant CSS can fold
  into that layout bundle
- `web/routes/dashboard/page.css` emits `routes/dashboard/page.css` only when no
  non-root layout owns that page
- `web/components/meter/meter.css` folds into whichever generated route/layout
  stylesheet imports `web/components/meter`

Empty or whitespace-only CSS still gets helper generation but is not injected
and does not write a browser file.

After route ownership is resolved, `gen assets` stages those generated
stylesheets and sends them through the same static asset builder as
`web/assets/**/*.css`. The final browser files are esbuild-transformed and
minified. This is a final CSS transform step; route/component CSS ownership is
still decided by `no-js`, and CSS does not emit shared esbuild chunks like
JavaScript.

Use colocated `.css` files when you need normal stylesheet features:

- pseudo-classes such as `:hover`, `:focus-visible`, or `:has(...)`
- pseudo-elements such as `::before`
- combinators such as `.card > .icon + .label`
- at-rules such as `@media`, `@container`, or `@scope`

Use templ `css {}` for simple component-scoped class declarations.

## Generated Scripts

Client Asset scripts may use these source extensions:

```text
.js
.ts
.tsx
.mjs
.mts
```

All script sources emit browser module entries with `.js` output paths that
mirror the owner source path. For example, `web/routes/dashboard/page.tsx`
emits `routes/dashboard/page.js`, and `web/components/meter/meter.ts` emits
`components/meter/meter.js`. Choose one source extension for each route owner
or component package. The validator enforces this because `page.ts` and
`page.tsx` both emit `page.js`, and `meter.ts` and `meter.tsx` both emit
`meter.js`.

Generated routes inject matched owner entries through `@metagen.Head(meta)`.
Normal pages do not need to call generated script helpers manually. Shared
esbuild chunks are imported by those owner entries, so generated metadata does
not list chunk files directly.

Client Asset scripts are bundled together with esbuild splitting enabled. Relative
imports and package imports are supported when esbuild can resolve them from the
app workspace:

```ts
import { animate } from "./animation";
import { createFocusTrap } from "focus-trap";
```

Package imports require the dependency to be installed in the app workspace.
TSX is bundled by esbuild, but `no-js` does not install or configure a JSX
runtime for you. Configure the app the same way you would for esbuild: provide
the JSX runtime imports, `jsxImportSource`, or related app settings your TSX
source expects. TypeScript and TSX are bundled, not typechecked. Run
`tsc --noEmit` in app validation when you need typechecking.

## Explicit `web/assets`

Use `web/assets` for files that need their own hashed URL and should not be
automatically attached to a route or component.

Good fits:

- images and fonts
- Open Graph images
- downloads
- embeddable CSS or JavaScript for another site
- vendor files
- a global stylesheet you include intentionally

Files under `web/assets`:

- are written under `web/assets-build`
- are fingerprinted in the manifest
- may be minified when they are `.css`, `.js`, `.mjs`, or `.cjs`
- are not auto-injected into pages
- do not get generated CSS constants or script helpers
- are not bundled as an import graph

Package imports from `node_modules` are not supported in `web/assets` CSS or
JavaScript. Browser-resolvable relative imports can work when the imported file
is also present under `web/assets`, but `no-js` does not rewrite import paths.

Reference a hashed asset intentionally from app code:

```go
metagen.AssetURL(ctx, "site.css")
```

## Fixed `web/public` Files

Use `web/public` when the request path must stay fixed:

```text
web/public/favicon.ico      -> /favicon.ico
web/public/site.webmanifest -> /site.webmanifest
```

Do not put files in `web/public` when they should be fingerprinted.

## Runtime Output

Generated Client Asset files, shared chunks, and explicit `web/assets` files
share the same runtime output directory:

```text
web/assets-build/
  manifest.json
  routes/root.css
  routes/layout.css
  routes/dashboard/layout.css
  routes/layout.js
  routes/dashboard/page.js
  chunks/chunk-ABC123.js
  styles/templ.css
  embed.js
```

At runtime, `httpserver.NewApp(...)` reads the manifest and serves files under:

```text
/_assets/<hash>/
```

The default manifest path is:

```text
web/assets-build/manifest.json
```

Override it only when you also configure the same path in runtime setup. See
[Bundle Config Reference](bundle-config.md) for build-time path settings.

## Common Fixes

### Route Asset Without Matching Template

A file like this fails validation:

```text
web/routes/dashboard/page.css
```

unless this file exists beside it:

```text
web/routes/dashboard/page.templ
```

Fix by adding the matching template, renaming the asset stem, or moving the file
to `web/assets` if it is not route-owned.

### Component Asset With The Wrong Stem

This fails:

```text
web/components/meter/theme.css
web/components/meter/behavior.ts
```

Use the component anchor stem instead:

```text
web/components/meter/meter.css
web/components/meter/meter.tsx
```

### Multiple Route Or Component Script Sources

This route fails because both files belong to `page.templ` and emit `page.js`:

```text
web/routes/dashboard/page.ts
web/routes/dashboard/page.tsx
```

This component fails because both files emit `meter.js`:

```text
web/components/meter/meter.ts
web/components/meter/meter.tsx
```

Keep one script source and move shared code into an imported support module
outside the component asset slot.

### Component File In The Root

This fails:

```text
web/components/badge.templ
```

Use a component package directory:

```text
web/components/badge/badge.templ
```

### Need A Global File On Every Page

If the file is generated by templ `css {}`, keep the default templ CSS pipeline.
If it is a handwritten stylesheet or script, put it in `web/assets` and include
it intentionally from metadata or your root template.

## Related Docs

- [Static Assets And Client Assets](../features/static-assets.md)
- [App Conventions](../conventions.md)
- [CLI Reference](cli.md)
- [Bundle Config Reference](bundle-config.md)
- [Metadata and Head](../features/metadata-and-head.md)
