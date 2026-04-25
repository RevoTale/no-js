# App Migrations

Use this page when upgrading an app between `no-js` versions.

Migration guides are ordered with the highest target version first. Files under
`docs/app/migrations/` are named by source and target version. A `next` target is
temporary and must be renamed to the concrete release version after release.

## Next

- [v1.3.0 to Next: View Package, Typed System Pages, Server Error UI, And Templ CSS Config](migrations/v1.3.0-to-next.md)
  Rename app `web/view` packages to `package view`, move 404 model construction
  into generated `Resolve...NotFound(...)` resolver methods, move custom 500 UI
  to `httpserver.CustomConfig.ServerErrorPage`, and remove `-templ-css` now that global templ CSS extraction is enabled by
  default and disabled with `assets.templ_css: false`.
