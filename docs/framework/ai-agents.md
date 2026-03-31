# AI Agents

This guide is for agents modifying the `no-js` framework repository itself.

## Read Order

When you need context, inspect in this order:

1. `README.md`
2. `docs/framework/developing-no-js.md`
3. `AGENTS.md`
4. `framework/httpserver/`
5. `framework/discovery/`
6. `internal/projectlayout/`
7. `internal/bundler/approutegen/`

If the task affects the consuming-app contract, also check:

- `docs/app/getting-started.md`
- `docs/app/conventions.md`

## Editing Rules

- Keep public framework config generic.
- Keep app-specific concepts out of public framework packages.
- Do not manually edit generated output.
- Surface wiring problems at startup, not through request-time soft fallbacks.
- Preserve the strict `web/*` app contract unless the task is explicitly about
  changing it.
- Treat `no-js.bundle.yaml` as build-time only.

## Change Routing

- HTTP handler assembly:
  `framework/httpserver`
- `robots.txt`, sitemap, or RSS behavior:
  `framework/discovery`
- route discovery or generated bundle behavior:
  `internal/bundler/approutegen`
- static asset generation:
  `internal/bundler/staticassets` and `framework/staticassets`
- app layout resolution:
  `internal/projectlayout`

## Validation

```bash
task fix
task validate
task test
```
