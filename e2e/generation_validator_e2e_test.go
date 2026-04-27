package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerationValidatorAllowsStrictRouteAndComponentShape(t *testing.T) {
	appDir := prepareFixtureAppWithNoJSGen(t, "clientassetsapp")

	_, err := os.Stat(filepath.Join(appDir, "web", "routes", "error.templ"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(appDir, "web", "routes", "page.css_gen.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(appDir, "web", "components", "meter", "meter.css_gen.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(appDir, "web", "components", "filterpanel", "filterpanel.css_gen.go"))
	require.NoError(t, err)
}

func TestGenerationValidatorAllowsSlotFallbackClientAssets(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	slotDir := filepath.Join(appDir, "web", "routes", "section", "_slot__aside")
	require.NoError(t, os.MkdirAll(slotDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(slotDir, "default.templ"), []byte("package aside\n"), 0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(slotDir, "default.css"),
		[]byte(".slot-fallback-e2e { --slot-default-fallback-e2e: 1; }\n"),
		0o644,
	))

	runGo(t, appDir, "tool", "no-js", "gen", "assets", "-root", ".")

	require.NoFileExists(
		t,
		filepath.Join(appDir, "web", "assets-build", "routes", "section", "_slot__aside", "default.css"),
	)
	sectionCSS, err := os.ReadFile(filepath.Join(appDir, "web", "assets-build", "routes", "section", "layout.css"))
	require.NoError(t, err)
	require.Contains(t, string(sectionCSS), "--slot-default-fallback-e2e")
}

func TestGenerationValidatorRejectsUnsupportedRouteFiles(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "routes", "helper.go"),
		[]byte("package routes\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `validate app shape: unsupported file in web/routes: "helper.go"`)
}

func TestGenerationValidatorRejectsRouteAssetWithoutMatchingTemplate(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	orphanDir := filepath.Join(appDir, "web", "routes", "orphan")
	require.NoError(t, os.MkdirAll(orphanDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(orphanDir, "page.css"), []byte(".orphan {}\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(
		t,
		string(output),
		`route Client Asset "orphan/page.css" requires matching template "orphan/page.templ"`,
	)
}

func TestGenerationValidatorRejectsSlotFallbackAssetOutsideSlot(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "routes", "default.css"),
		[]byte(".fallback {}\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "assets", "-root", ".")
	require.Contains(t, string(output), `default Client Assets are only allowed inside slot directories`)
}

func TestGenerationValidatorRejectsSlotFallbackAssetBelowSlotRoot(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	slotNestedDir := filepath.Join(appDir, "web", "routes", "dashboard", "_slot__aside", "nested")
	require.NoError(t, os.MkdirAll(slotNestedDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(slotNestedDir, "default.css"), []byte(".fallback {}\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "assets", "-root", ".")
	require.Contains(t, string(output), `default Client Assets are only allowed at the slot root`)
}

func TestGenerationValidatorRejectsSlotFallbackWithoutOwnerLayout(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	slotDir := filepath.Join(appDir, "web", "routes", "orphan", "_slot__aside")
	require.NoError(t, os.MkdirAll(slotDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(slotDir, "default.templ"), []byte("package aside\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(slotDir, "default.css"), []byte(".fallback {}\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "assets", "-root", ".")
	require.Contains(t, string(output), `requires owner layout "orphan/layout.templ"`)
}

func TestGenerationValidatorRejectsPageAndMethodRouteConflict(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "web", "routes", "route.go"), []byte("package routes\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "assets", "-root", ".")
	require.Contains(t, string(output), `route pattern conflict`)
	require.Contains(t, string(output), `both resolve to "/"`)
}

func TestGenerationValidatorRejectsMultipleRouteScriptSources(t *testing.T) {
	appDir := copyFixtureApp(t, "routepagecssapp")
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "routes", "page.ts"),
		[]byte("console.log('page')\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "routes", "page.tsx"),
		[]byte("console.log('page')\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `route directory web/routes has multiple script sources for "page.templ"`)
	require.Contains(t, string(output), `choose one of page.js, page.ts, page.tsx, page.mjs, or page.mts`)
}

func TestGenerationValidatorRejectsUnsupportedComponentFiles(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "components", "meter", "logo.svg"),
		[]byte("<svg/>\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `validate app shape: unsupported file in web/components: "meter/logo.svg"`)
}

func TestGenerationValidatorRejectsLooseComponentRootFile(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	require.NoError(t, os.WriteFile(
		filepath.Join(appDir, "web", "components", "badge.templ"),
		[]byte("package components\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `validate app shape: unsupported file in web/components: "badge.templ"`)
	require.Contains(t, string(output), `web/components/badge/badge.templ`)
}

func TestGenerationValidatorRejectsExtraComponentTemplate(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	cardDir := filepath.Join(appDir, "web", "components", "card")
	require.NoError(t, os.MkdirAll(cardDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "card.templ"), []byte("package card\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "variants.templ"), []byte("package card\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `validate app shape: unsupported component template "card/variants.templ"`)
	require.Contains(t, string(output), `card.templ`)
}

func TestGenerationValidatorRejectsMultipleComponentScriptSources(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	cardDir := filepath.Join(appDir, "web", "components", "card")
	require.NoError(t, os.MkdirAll(cardDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "card.templ"), []byte("package card\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "card.ts"), []byte("console.log('card')\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "card.tsx"), []byte("console.log('card')\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `component package web/components/card has multiple script sources`)
	require.Contains(t, string(output), `choose one of card.js, card.ts, card.tsx, card.mjs, or card.mts`)
}

func TestGenerationValidatorRejectsExportedComponentSupportGo(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	cardDir := filepath.Join(appDir, "web", "components", "card")
	require.NoError(t, os.MkdirAll(cardDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cardDir, "card.templ"), []byte("package card\n"), 0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(cardDir, "helper.go"),
		[]byte("package card\n\nfunc Helper() {}\n"),
		0o644,
	))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `exported declaration "Helper" in web/components/card/helper.go`)
	require.Contains(t, string(output), `must move to "card.go" or be made private`)
}

func TestGenerationValidatorRejectsComponentAssetFolders(t *testing.T) {
	appDir := copyFixtureApp(t, "clientassetsapp")
	themeDir := filepath.Join(appDir, "web", "components", "theme")
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "theme.css"), []byte(".theme {}\n"), 0o644))

	output := runNoJSError(t, appDir, "gen", "routes", "-root", ".")
	require.Contains(t, string(output), `component package web/components/theme must contain "theme.templ" or "theme.go"`)
}
