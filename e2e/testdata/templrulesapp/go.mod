module example.com/no-js-e2e/templrulesapp

go 1.25.0

require (
	github.com/RevoTale/no-js v1.3.0
	github.com/a-h/templ v0.3.1001
)

require (
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/evanw/esbuild v0.28.0 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.6.1 // indirect
	github.com/tdewolff/parse/v2 v2.8.12 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	github.com/RevoTale/no-js/cmd/no-js
	github.com/RevoTale/no-js/cmd/templgen
)

replace github.com/RevoTale/no-js => ../../..
