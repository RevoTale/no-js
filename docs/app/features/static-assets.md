# Static Assets And Client Assets

Use this guide when you need CSS, JavaScript, TypeScript, images, downloads, or
fixed public files in a `no-js` app.

Start with templ `css {}` for simple component-scoped styles. Use Client
Assets for colocated CSS files and scripts next to the route, layout, 404 page,
or component that uses them. Use `web/assets` only when a file must be
addressed as its own hashed URL.

Your root template must render `@metagen.Head(meta)`. That is where `no-js`
adds managed route stylesheets and module scripts.

## Default Asset Shape

Most apps use all three asset lanes, but each lane has a different job:

- templ `css {}`
  Use for simple component-scoped class styles. `no-js` extracts these into one
  global `styles/templ.css` by default when declarations exist.
- Client Assets
  Use for route, layout, 404, or component CSS/JS/TS. CSS is folded into root,
  layout-subtree, or page-fallback stylesheets. Scripts stay as owner module
  entries, with shared imports emitted as esbuild chunks.
- `web/assets`
  Use for images, fonts, downloads, vendor files, or files consumed outside the
  route graph. `no-js` fingerprints them under `/_assets/<hash>/`, but app
  code must reference them intentionally.

`go tool no-js gen assets` sends every final `.css` file through the static
asset builder. Global templ CSS, generated route/component Client Asset
stylesheets, and explicit `web/assets/**/*.css` files are transformed and
minified by esbuild before the browser receives them. Dynamic templ `css {}`
that stays local to rendering is not part of that global asset build.

The final asset pipeline is:

```text
templ global css -> staged styles/templ.css -> static esbuild build -> hashed CSS
route/component css -> no-js ownership/class rewrite -> staged CSS -> static esbuild build -> hashed CSS
client scripts -> esbuild owner entries/chunks -> static hash -> hashed JS
web/assets css/js -> static esbuild build -> hashed asset
```

All esbuild paths use `assets.browser_targets`. The default is `es2020`. Add
browser-specific targets only when the app needs older syntax output or CSS
vendor prefixes:

```yaml
version: 1

assets:
  browser_targets:
    - es2020
    - safari13
```

You do not need `no-js.bundle.yaml` for the default templ CSS behavior. Add the
file only when you need to change a default, such as disabling global templ CSS
extraction or changing browser targets.

## Choose Templ CSS Or CSS Files

Prefer templ `css {}` components for simple component-scoped class styles:

```templ
css meterRoot() {
	padding: 16px;
}

templ Meter(label string) {
	<section class={ meterRoot() }>
		{ label }
	</section>
}
```

Use colocated `.css` files when you need normal stylesheet selectors or larger
CSS composition:

- pseudo-classes such as `:hover`, `:focus-visible`, or `:has(...)`
- pseudo-elements such as `::before`
- combinators such as `.card > .icon + .label`
- at-rules such as `@media`, `@container`, or `@scope`
- route, layout, 404, or component styles that are easier to read as a
  stylesheet

## Pick The Right Place

Use the exact owner stem in route and component Client Asset filenames:

| You need | Put it here |
| --- | --- |
| Simple component-scoped class CSS | templ `css {}` in the `.templ` file |
| Shell CSS for every page | `web/routes/root.css` beside `root.templ` |
| CSS for a route page | `web/routes/dashboard/page.css` beside `page.templ` |
| CSS for a layout | `web/routes/dashboard/layout.css` beside `layout.templ` |
| CSS for a slot fallback | `web/routes/dashboard/_slot__aside/default.css` beside `default.templ` |
| CSS for a 404 page | `web/routes/dashboard/404.css` beside `404.templ` |
| CSS for a component | `web/components/meter/meter.css` beside `meter.templ` or `meter.go` |
| JS/TS/TSX for a route page | `web/routes/dashboard/page.tsx` beside `page.templ` |
| JS/TS/TSX for a layout | `web/routes/dashboard/layout.tsx` beside `layout.templ` |
| JS/TS/TSX for a 404 page | `web/routes/dashboard/404.tsx` beside `404.templ` |
| JS/TS/TSX for a component | `web/components/meter/meter.tsx` beside `meter.templ` or `meter.go` |
| CSS or JS consumed by another website | `web/assets/embed.css` or `web/assets/embed.js` |
| Fonts, images, downloads, Open Graph images, vendor files | `web/assets` |
| Fixed paths like `/favicon.ico` or `/site.webmanifest` | `web/public` |

## Add Route CSS

Create a CSS file beside the route template:

```text
web/routes/dashboard/page.templ
web/routes/dashboard/page.css
```

Use the same file stem as the route template. For example, `page.templ` owns
`page.css`, `layout.templ` owns `layout.css`, slot-root `default.templ` owns
`default.css`, and `404.templ` owns `404.css`.

```css
/* web/routes/dashboard/page.css */
.shell {
	padding: 16px;
}
```

Use the generated class constant from templ:

```templ
package dashboard

templ Page(model view.DashboardPageView) {
	<main class={ PageShellClass }>
		{ model.Title }
	</main>
}
```

Run generation:

```bash
go tool no-js gen -root .
```

`no-js` generates `page.css_gen.go` and anonymizes the class name in rendered
HTML. The browser stylesheet depends on the route shape:

- if the page is under a non-root `layout.templ`, the page CSS is folded into
  that layout subtree stylesheet, such as `routes/dashboard/layout.css`
- if there is no non-root layout owner, the page gets a fallback stylesheet,
  such as `routes/dashboard/page.css`
- `web/routes/root.css` is shell CSS and stays separate from page/layout CSS

Generated routes inject the needed stylesheets through `@metagen.Head(meta)`.

## Add Component CSS

Put component CSS beside the component:

```text
web/components/meter/meter.templ
web/components/meter/meter.css
```

```templ
package meter

templ Meter(label string) {
	<section class={ MeterRootClass }>
		{ label }
	</section>
}
```

When a route or layout imports `web/components/meter`, `meter.css` is folded
into the stylesheet owned by the nearest non-root layout. If there is no
non-root layout, it is folded into the page fallback stylesheet. Routes outside
that layout subtree do not receive the component CSS.

Component packages are package-owned for Client Assets, and asset names must
match the component package anchor. For `web/components/meter`, use
`meter.css` plus at most one script source: `meter.js`, `meter.ts`, `meter.tsx`,
`meter.mjs`, or `meter.mts`. Do not add `theme.css`, `behavior.ts`, or more
than one same-stem script source inside that package. Other static files, such as
images, fonts, downloads, docs, or JSON data, do not belong in
`web/components`; put them under `web/assets`, `web/public`, or another
app-owned package.

## Add Route Scripts

Create a script beside the route template:

```text
web/routes/dashboard/page.ts
```

Use the same file stem as the route template: `page.ts` for `page.templ`,
`layout.ts` for `layout.templ`, and `404.ts` for `404.templ`. Choose one script
extension per route owner; `page.ts` and `page.tsx` both emit `page.js`, so they
cannot live beside the same `page.templ`.

```ts
import { createFocusTrap } from "focus-trap";

const trap = createFocusTrap("[data-dashboard-panel]");
trap.activate();
```

Run generation:

```bash
go tool no-js gen -root .
```

The matched route receives the owner module entries it needs. Scripts are not
folded into layout CSS bundles:

```html
<script type="module" src="/_assets/<hash>/routes/dashboard/page.js"></script>
```

Normal pages should not call script helpers manually. Generated routes add
owner scripts once through `@metagen.Head(meta)`. If a layout or imported
component is reused by many routes, all those routes point to the same owner
script path. Shared imports are emitted as esbuild
chunks and loaded by the owner scripts.

## JavaScript And TypeScript Imports

Client Asset scripts are bundled together with esbuild splitting enabled.
Relative imports and package imports are supported:

```ts
import { animate } from "./animation";
import { createFocusTrap } from "focus-trap";
```

Package imports require the dependency to be installed in the app workspace.
`no-js` does not install `node_modules` for you.

TSX is bundled by esbuild, but `no-js` does not configure React, Preact, or any
other JSX runtime for you. Provide the imports, `jsxImportSource`, or app-level
settings your TSX source expects. TypeScript and TSX are bundled, not
typechecked. Keep `tsc --noEmit` in your app validation flow when you need
typechecking.

## CSS Imports

Do not use CSS `@import` as a route/component dependency system:

```css
@import "some-package/styles.css";
```

Client Asset CSS is discovered from route templates and imported Go component
packages, not from CSS imports. Put CSS beside the route, layout, 404 page, or
component that owns it.

Explicit `web/assets` CSS is different. A CSS file under `web/assets` can import
another browser-resolvable CSS file with a relative path, and asset generation
bundles it into the importing CSS file:

```css
@import "./reset.css";
```

Use that for manual global files you include intentionally from metadata or
app-owned head code. Package imports from `node_modules` are not supported in
CSS; vendor the CSS under `web/assets` or use the package from a Client Asset
script when the dependency is JavaScript.

## Use `web/assets` For Explicit Hashed Files

Use `web/assets` when a file needs its own hashed URL outside the route graph.

Good fits:

- an embeddable JavaScript module for external websites
- fonts or images referenced from CSS
- Open Graph images
- downloads
- vendor files
- a global stylesheet you include yourself

Example:

```text
web/assets/embed.js
web/assets/fonts/brand.woff2
web/assets/site.css
web/assets/shared/reset.css
```

A manual stylesheet can import another CSS file under `web/assets`:

```css
/* web/assets/site.css */
@import "./shared/reset.css";

.site-shell {
	color: #2563eb;
}
```

The generated `site.css` URL points to one hashed CSS file that already contains
`shared/reset.css` before `.site-shell`. Include it intentionally from metadata
or app-owned head code; it is not attached to routes automatically.

Run:

```bash
go tool no-js gen assets -root .
```

The files are written under `web/assets-build` and served from:

```text
/_assets/<hash>/embed.js
/_assets/<hash>/fonts/brand.woff2
```

`web/assets` files are not scoped, do not get generated class constants, and are
not auto-injected into pages. CSS files may bundle browser-resolvable relative
`@import` files; JavaScript files are minified as standalone files, not bundled
as a module graph.

If you need to add one to a page, add the URL yourself. Metadata resolvers can
use the request context to resolve hashed asset URLs:

```go
func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Stylesheets: []string{
			metagen.AssetURL(meta.Context(), "site.css"),
		},
	}, nil
}
```

## `web/assets` Imports

CSS files in `web/assets` are built with esbuild in bundle mode. Relative CSS
imports are bundled into the importing file when the browser path can be
resolved from `web/assets`. Relative `url(...)` references inside imported CSS
are rebased so they still point at the correct copied asset from the final
bundled CSS file:

```css
@import "./reset.css";
```

JavaScript files in `web/assets` are minified and fingerprinted as standalone
files. They are not bundled as a module graph, and package imports from
`node_modules` are not supported there. Use Client Asset scripts when you need
esbuild package resolution, TypeScript, TSX, or shared JS chunks.

## Use `web/public` For Fixed Paths

Use `web/public` when the request path must stay fixed:

```text
web/public/favicon.ico        -> /favicon.ico
web/public/site.webmanifest   -> /site.webmanifest
```

Do not put files in `web/public` when they should be fingerprinted.

## Bundle Templ CSS

templ `css {}` components are a good default for simple component-scoped
classes. By default, `no-js` extracts them into the hashed asset path when
zero-argument templ CSS declarations or `TemplCSSVariants()` exist.

```bash
go tool no-js gen assets -root .
```

That collects registered templ CSS classes into one global stylesheet,
`styles/templ.css`, and sends it through the hashed asset path. When the
stylesheet exists, `httpserver.NewApp(...)` injects it into managed head output
for every page and suppresses the registered inline templ CSS.

Use colocated `.css` files instead when the style needs route-level delivery,
selectors, pseudo classes, pseudo elements, combinators, or other
stylesheet-level composition.

To keep templ CSS on the templ inline registration path, opt out in bundle
config:

```yaml
version: 1

assets:
  templ_css: false
```

If you need parameterized templ CSS variants, return them from
`TemplCSSVariants()` in `web/view`. If you only use zero-argument templ `css`
components, omit the hook.

## Related Docs

- [Asset Pipeline Reference](../reference/asset-pipeline.md)
- [Metadata and Head](metadata-and-head.md)
- [CLI Reference](../reference/cli.md)
- [Bundle Config Reference](../reference/bundle-config.md)
