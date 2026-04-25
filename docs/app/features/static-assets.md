# Static Assets And Client Assets

Use this guide when you need CSS, JavaScript, TypeScript, images, downloads, or
fixed public files in a `no-js` app.

Start with templ `css {}` for simple component-scoped styles. Use Client
Assets for colocated CSS files and scripts next to the route, layout, 404 page,
or component that uses them. Use `web/assets` only when a file must be
addressed as its own hashed URL.

Your root template must render `@metagen.Head(meta)`. That is where `no-js`
adds managed route stylesheets and module scripts.

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

| You need | Put it here |
| --- | --- |
| Simple component-scoped class CSS | templ `css {}` in the `.templ` file |
| CSS for a route, layout, or 404 page | same-stem `.css` beside the route template, such as `page.css` beside `page.templ` |
| CSS for a component | `.css` in the component package |
| JS/TS for a route, layout, or 404 page | same-stem `.js`, `.ts`, `.mjs`, or `.mts` beside the route template |
| JS/TS for a component | `.js`, `.ts`, `.mjs`, or `.mts` in the component package |
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
`page.css`, `layout.templ` owns `layout.css`, and `404.templ` owns `404.css`.

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

`no-js` generates `page.css_gen.go`, anonymizes the class name in rendered HTML,
and injects the route stylesheet through `@metagen.Head(meta)`.

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

When a route imports `web/components/meter`, that route receives the component
CSS once. Routes that do not import the component do not receive that CSS.

Component packages are package-owned for Client Assets. You can keep component
CSS and scripts in the component package; they do not need to match the template
file stem.

## Add Route Scripts

Create a script beside the route template:

```text
web/routes/dashboard/page.ts
```

Use the same file stem as the route template: `page.ts` for `page.templ`,
`layout.ts` for `layout.templ`, and `404.ts` for `404.templ`.

```ts
import { createFocusTrap } from "focus-trap";

const trap = createFocusTrap("[data-dashboard-panel]");
trap.activate();
```

Run generation:

```bash
go tool no-js gen -root .
```

The matched route receives one module script:

```html
<script type="module" src="/_assets/<hash>/routes/dashboard.js"></script>
```

Normal pages should not call script helpers manually. Generated routes add
route scripts once through `@metagen.Head(meta)`.

## JavaScript And TypeScript Imports

Client Asset scripts are bundled with esbuild. Relative imports and package
imports are supported:

```ts
import { animate } from "./animation";
import { createFocusTrap } from "focus-trap";
```

Package imports require the dependency to be installed in the app workspace.
`no-js` does not install `node_modules` for you.

TypeScript is bundled, not typechecked. If your app uses TypeScript, keep
`tsc --noEmit` in your app validation flow.

## CSS Imports

Do not use CSS `@import` as a route/component dependency system:

```css
@import "some-package/styles.css";
```

Client Asset CSS is discovered from route files and imported Go component
packages, not from CSS imports. Put CSS beside the route or component that owns
it.

If you need a third-party CSS file as an independent hashed file, put it under
`web/assets` and include it intentionally.

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
```

Run:

```bash
go tool no-js gen assets -root .
```

The files are written under `web/assets-build` and served from:

```text
/_assets/<hash>/embed.js
/_assets/<hash>/fonts/brand.woff2
```

`web/assets` files are not scoped, do not get generated class constants, are not
auto-injected into pages, and are not bundled as a module graph.

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

CSS and JavaScript files in `web/assets` may be minified and fingerprinted, but
imports are not resolved or rewritten.

Browser-resolvable relative imports can work when the imported file is also
present under `web/assets`:

```css
@import "./reset.css";
```

Package imports from `node_modules` are not supported in `web/assets` CSS or
JavaScript. Use Client Asset scripts when you need esbuild package resolution.

## Use `web/public` For Fixed Paths

Use `web/public` when the request path must stay fixed:

```text
web/public/favicon.ico        -> /favicon.ico
web/public/site.webmanifest   -> /site.webmanifest
```

Do not put files in `web/public` when they should be fingerprinted.

## Bundle Templ CSS

templ `css {}` components are a good default for simple component-scoped
classes. If you want `no-js` to extract them into the hashed asset path, run:

```bash
go tool no-js gen assets -root . -templ-css
```

That collects registered templ CSS classes into one global stylesheet,
`styles/templ.css`, and sends it through the hashed asset path. When the
stylesheet exists, `httpserver.NewApp(...)` injects it into managed head output
for every page and suppresses the registered inline templ CSS.

Use colocated `.css` files instead when the style needs route-level delivery,
selectors, pseudo classes, pseudo elements, combinators, or other
stylesheet-level composition.

If you need parameterized templ CSS variants, return them from
`TemplCSSVariants()` in `web/view`. If you only use zero-argument templ `css`
components, omit the hook.

## Related Docs

- [Asset Pipeline Reference](../reference/asset-pipeline.md)
- [Metadata and Head](metadata-and-head.md)
- [CLI Reference](../reference/cli.md)
- [Bundle Config Reference](../reference/bundle-config.md)
