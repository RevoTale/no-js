# AI Agents

This guide is for agents modifying the `no-js` framework repository itself.

## Read Order

When you need context, inspect in this order:

1. `README.md`
2. `docs/app/overview.md`
3. `docs/framework/developing-no-js.md`
4. `AGENTS.md`
5. `framework/httpserver/`
6. `framework/discovery/`
7. `internal/projectlayout/`
8. `internal/bundler/approutegen/`
9. `internal/bundler/clientassets/`

If the task affects the consuming-app contract, also check:

- `docs/app/reference/cli.md`
- `docs/app/reference/bundle-config.md`
- `docs/app/reference/httpserver.md`
- `docs/app/reference/asset-pipeline.md`
- `docs/app/getting-started.md`
- `docs/app/conventions.md`
- `docs/app/troubleshooting.md`

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
- Client Assets:
  `internal/bundler/clientassets` and `docs/framework/client-assets-pipeline.md`
- static asset manifest and fingerprinting:
  `internal/bundler/staticassets` and `framework/staticassets`
- app layout resolution:
  `internal/projectlayout`

## Validation

```bash
task fix
task validate
task test
```
