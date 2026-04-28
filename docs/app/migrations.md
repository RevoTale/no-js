# App Migrations

Use this page when upgrading an app between `no-js` versions.

Migration guides are ordered with the highest target version first. Files under
`docs/app/migrations/` are named by source and target version. A `next` target is
temporary and must be renamed to the concrete release version after release.

## Next

- [v1.3.0 to Next: Typed System Pages, 404 Metadata, Server Errors,
  Templ CSS, And App Shape](migrations/v1.3.0-to-next.md)
  Rename app `web/view` packages to `package view`, move 404 model construction
  into generated `Resolve...NotFound(...)` resolver methods, add generated
  `MetaGen...NotFound(...)` methods for 404 head metadata, move custom 500 UI to
  `httpserver.CustomConfig.ServerErrorPage`, remove `-templ-css` now that global
  templ CSS extraction is enabled by default and disabled with
  `assets.templ_css: false`, and clean unsupported files from `web/routes` and
  `web/components`.
