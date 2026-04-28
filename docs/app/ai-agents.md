# AI Agents For App Development

This guide is for AI agents building a consuming app with `no-js`.

If you are changing the `no-js` framework itself, use
[Framework AI Agents](../framework/ai-agents.md) instead.

If the consuming app has its own `AGENTS.md`, read the nearest one first. Use
this page for the `no-js`-specific app contract.

## Mental Model

`no-js` owns generation and runtime glue. The app owns routes, view models,
resolver implementations, services, and policy.

The happy path is:

```go
handler, err := httpserver.NewApp(httpserver.Config[*view.Context]{
	App: generated.Bundle(appContext),
})
```

Treat generated code as a contract:

- read `web/resolvers/generated.go`
- implement the missing resolver methods
- regenerate after changing `web/routes`, metadata, i18n, or assets
- do not edit `web/generated/*` or generated resolver contracts by hand

## Sources Of Truth

- app shape: [App Conventions](conventions.md)
- first runnable app: [Getting Started](getting-started.md)
- generated method names: `web/resolvers/generated.go`
- build command: `go tool no-js gen -root .`
- extra templ packages outside `web/routes`: `go tool templgen`
- asset placement: [Static Assets](features/static-assets.md) and
  [Asset Pipeline Reference](reference/asset-pipeline.md)
- app workflow: the app's `Taskfile.yml` when it exists

## Agent Loop

1. Read the nearest app `AGENTS.md` if it exists.
2. Read `web/resolvers/generated.go` before implementing resolvers.
3. Change source files only: `web/routes`, `web/resolvers`, `web/view`,
   `web/components`, `internal`, app config, or app entrypoints.
4. Run the app workflow if present: `task gen`, `task validate`, `task test`.
5. If no Taskfile exists, run `go tool no-js gen -root .`, then app tests.
6. If generation changes resolver contracts, implement the new methods instead
   of editing generated files.
7. Before finishing, check that generated output is committed with the app.

## Directory Model

Use this shape for a new app:

```text
your-app/
  go.mod
  cmd/
    server/
      main.go
  internal/
    <domain>/
      service.go
      repository.go
  web/
    routes/
      root.templ
      page.templ
      404.templ
    generated/
    components/
      card/
        card.templ
    assets/
      embed.js
    public/
      favicon.ico
    resolvers/
    view/
      context.go
      view_models.go
```

Meaning:

- `cmd/server`: process entrypoint and HTTP server startup
- `internal`: app domain code, repositories, API clients, mailers, jobs, and
  other code not directly about rendering
- `web/routes`: route templates, layouts, 404 templates, slots, and route-local
  assets
- `web/generated`: generated output; commit it, but do not edit it by hand
- `web/resolvers`: request-to-view-model glue generated contracts ask for
- `web/view`: app context and display-specific view models
- `web/components`: reusable rendering components, one package per component
- `web/assets` and `web/public`: global hashed assets and fixed-path public
  files

Go `internal` is intentional. Packages inside `internal` can be imported only by
code under the parent tree, which makes it a good place for app-only services
that should not become rendering API.

Use the generated `Resolver` type from `web/resolvers/generated.go`. In the
default bundle path the resolver is `route_resolvers.Resolver{}`, so app
services should usually be reachable from `*view.Context` or from types it owns.
Keep process wiring in `cmd/server`.

## Working Rules

Do:

- keep templates focused on rendering and light presentation logic
- shape display data in Go before render
- use explicit view models for pages and components
- pass ordinary render data as templ parameters
- use implicit templ `ctx` only for cross-cutting request context
- put metadata in generated `MetaGen*` resolver methods
- use stable `data-*`, `id`, or `x-ref` hooks for JavaScript
- use `templ.URL(...)` for dynamic `hx-*` or other non-standard URL attributes
- keep domain entities data-focused; put repositories, mailers, and API clients
  in services
- use the app's `task gen`, `task validate`, or `task test` when those tasks
  exist
- run `go tool no-js gen -root .` after app shape changes
- run app tests after generation

Do not:

- call repositories, mailers, external APIs, or services from `.templ` files
- hide domain logic in `web/routes`
- query templ-generated CSS class names from JavaScript
- manually inject route assets that generation already manages
- add arbitrary files directly under `web/components`
- hand-edit generated files to make compilation pass
- invent resolver method signatures; copy them from `web/resolvers/generated.go`

## Client Assets Loop

Route and component CSS/JS are generation inputs. Add files beside the owner
that uses them, then regenerate.

```text
web/routes/products/page.templ
web/routes/products/page.css
web/routes/products/page.ts

web/components/card/card.templ
web/components/card/card.css
web/components/card/card.ts
```

Rules:

- owner assets use the same stem as the owner template: `page.css`,
  `layout.tsx`, `404.css`, `card.ts`
- each route or component owner has at most one script source extension
- imported component assets are discovered from component imports in templates
- route/component CSS imports do not define asset ownership
- TSX is bundled, but `no-js` does not configure a JSX runtime or typecheck it
- `web/assets` is for explicit global hashed files, not normal route CSS/JS
- `web/public` is for fixed-path public files

Use `web/assets` when a file needs its own hashed URL outside the route graph:

- embed CSS or JS consumed by another website
- global vendor CSS or JS that the app intentionally injects
- fonts, images, downloads, Open Graph images, and other addressable files

Do not use `web/assets` for normal page, layout, 404, or component CSS/JS. Files
under `web/assets` are not auto-injected and do not get generated class
constants or script helpers. Reference them intentionally, for example from
metadata:

```go
return metagen.Metadata{
	Stylesheets: []string{
		metagen.AssetURL(meta.Context(), "site.css"),
	},
}, nil
```

Run:

```bash
go tool no-js gen -root .
```

That writes route helpers, Client Asset helpers, and bundled browser assets. If
the app splits the workflow, `go tool no-js gen routes -root .` writes generated
Go helpers and source-adjacent Client Asset helpers, while
`go tool no-js gen assets -root .` writes the browser CSS/JS bundles.

## Patterns And Anti-patterns

The examples below use generated root-route method shapes. App fields such as
`Catalog` and view-model helpers are illustrative; generated method names and
parameters are not. Each pair shows the same task with one important decision
changed.

### Resolver Owns Data Loading

Load service data in the resolver and pass render-ready data to the template.

#### Pattern

```go
func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	_ *http.Request,
	_ RootParams,
) (view.RootPageView, error) {
	products, err := appCtx.Catalog.Featured(ctx)
	if err != nil {
		return view.RootPageView{}, err
	}
	return view.RootPageView{Products: view.ProductCards(products)}, nil
}
```

```templ
templ Page(model view.RootPageView) {
	<ul>
		for _, product := range model.Products {
			<li>{ product.Name }</li>
		}
	</ul>
}
```

The pattern keeps data loading and error handling in Go before render.

#### Anti-pattern

```templ
templ Page(app *view.Context) {
	<ul>
		for _, product := range app.Catalog.MustFeatured(ctx) {
			<li>{ product.Name }</li>
		}
	</ul>
}
```

The anti-pattern performs service work from the template.

### Templates Own Markup

Templates should render a view model. They should not perform application
actions.

#### Pattern

```go
type CheckoutPageView struct {
	Paid bool
}
```

```templ
templ CheckoutPage(model view.CheckoutPageView) {
	<main>
		if model.Paid {
			<p>Paid</p>
		}
	</main>
}
```

The pattern gives the template one display decision that already came from app
code.

#### Anti-pattern

```templ
templ CheckoutPage(checkout *domain.Checkout) {
	if checkout.Charge(ctx) {
		<main>Paid</main>
	}
}
```

The anti-pattern charges during rendering instead of rendering existing state.

### Metadata Owns Head Data

Set page head data in generated metadata resolvers.

#### Pattern

```go
func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	_ RootParams,
) (metagen.Metadata, error) {
	alternates, err := meta.Alternates(meta.Locale(), nil)
	if err != nil {
		return metagen.Metadata{}, err
	}
	return metagen.Metadata{
		Title:      "Home",
		Alternates: alternates,
	}, nil
}
```

The pattern lets `@metagen.Head(meta)` compose title, canonical links,
alternates, and managed assets in the root layout.

#### Anti-pattern

```templ
templ Page(model view.RootPageView) {
	<head><title>{ model.Title }</title></head>
}
```

The anti-pattern puts document head data in a page template.

### Assets Follow Route Ownership

Colocate route assets with the route template that owns them.

#### Pattern

```text
web/routes/products/page.templ
web/routes/products/page.css
web/routes/products/page.ts
```

#### Anti-pattern

```text
web/routes/products/page.templ
web/routes/products/styles.css
web/routes/products/product.ts
```

The anti-pattern uses arbitrary asset names. Generation discovers route assets
by same-stem ownership, such as `page.css` and `page.ts` beside `page.templ`.
Put helper Go packages, images, data files, and domain code outside
`web/routes`.

### Components Are Packages

Each component is a Go package under `web/components/<name>/`.

#### Pattern

```text
web/components/card/card.templ
web/components/card/card.css
web/components/card/card.ts
```

#### Anti-pattern

```text
web/components/card.templ
web/components/card.css
web/components/card.ts
```

Component directory basename, Go package name, anchor file stem, and asset stem
must align. Files do not live directly under `web/components`.

### Explicit Assets Are Opt-in

Use `web/assets` for hashed files outside the generated route graph.

#### Pattern

```text
web/assets/embed.js
web/assets/site.css
web/assets/og/default.png
```

```go
func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	_ RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Stylesheets: []string{
			metagen.AssetURL(meta.Context(), "site.css"),
		},
	}, nil
}
```

The pattern makes global or embeddable assets explicit and hashed.

#### Anti-pattern

```text
web/assets/products.css
web/assets/card.ts
```

The anti-pattern uses `web/assets` for normal route or component assets. Put
those beside `page.templ`, `layout.templ`, `404.templ`, or the component anchor
instead.

### JavaScript Uses Stable Hooks

Use stable `data-*`, `id`, or `x-ref` hooks for client behavior.

#### Pattern

```templ
css saveButton() {
	display: inline-flex;
}

templ SaveButton(label string) {
	<button type="button" class={ saveButton() } data-save-button>{ label }</button>
}
```

```js
document.querySelector("[data-save-button]")?.addEventListener("click", save);
```

The pattern keeps generated CSS classes for styling and gives JavaScript a
stable hook.

#### Anti-pattern

```templ
css saveButton() {
	display: inline-flex;
}

templ SaveButton(label string) {
	<button type="button" class={ saveButton() }>{ label }</button>
}
```

```js
document.querySelector(".templ_123abc").addEventListener("click", save);
```

The anti-pattern queries a templ-generated class, which is not a cross-file
contract.

## Greenfield Flow

1. Create `go.mod`.
2. Add `github.com/RevoTale/no-js`, `github.com/a-h/templ`, and the `no-js`
   Go tool.
3. Create `cmd/server`, `web/routes`, `web/resolvers`, and `web/view`.
4. Add `root.templ`, one `page.templ`, and `404.templ`.
5. Run `go tool no-js gen -root .`.
6. Implement the generated resolver methods.
7. Wire `generated.Bundle(appContext)` into `httpserver.NewApp(...)`.
8. Add app services under `internal` when the page needs real data.
9. Run `go tool no-js gen -root .` again.
10. Run tests.

Prefer the full example in [Getting Started](getting-started.md) when creating
the first app files; this page explains agent behavior, not every file body.

## Read Next

1. [Getting Started](getting-started.md)
2. [App Conventions](conventions.md)
3. [Routing and Generation](features/routing-and-generation.md)
4. [Metadata and Head](features/metadata-and-head.md)
5. [Static Assets](features/static-assets.md)
6. [CLI Reference](reference/cli.md)
