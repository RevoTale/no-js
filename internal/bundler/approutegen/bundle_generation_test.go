package approutegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/internal/bundler/clientassets"
	"github.com/RevoTale/no-js/internal/bundler/viewcontract"
	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestGenerateBundleSourceWiresTemplCSSRegistryWhenEnabled(t *testing.T) {
	t.Parallel()

	source, err := generateBundleSource(projectlayout.ProjectLayout{
		AppModulePath:   testAppModulePath,
		GeneratedImport: "web/generated",
		ViewImport:      "web/view",
		Assets:          projectlayout.AssetsLayout{TemplCSS: true},
	}, zeroArgViewInspection())
	require.NoError(t, err)

	text := string(source)
	require.Contains(t, text, "func Bundle(appContext *view.Context) httpserver.AppBundle[*view.Context]")
	require.Contains(t, text, "resolvers := NewRouteResolvers()")
	require.Contains(t, text, "Handlers:                      Handlers(resolvers),")
	require.Contains(t, text, "NotFoundPage:                  NotFoundPage(resolvers),")
	require.Contains(t, text, "TemplCSSClasses:               TemplCSSClasses,")
	require.Contains(t, text, "OnStaticAssetBasePathResolved: nil,")
}

func TestGenerateBundleSourceDisablesTemplCSSRegistryWhenDisabled(t *testing.T) {
	t.Parallel()

	source, err := generateBundleSource(projectlayout.ProjectLayout{
		AppModulePath:   testAppModulePath,
		GeneratedImport: "web/generated",
		ViewImport:      "web/view",
	}, zeroArgViewInspection())
	require.NoError(t, err)

	text := string(source)
	require.Contains(t, text, "TemplCSSClasses:               nil,")
	require.NotContains(t, text, "TemplCSSClasses:               TemplCSSClasses,")
}

func TestClientAssetRouteSpecsAttachSlotTemplatesToOwnerRoute(t *testing.T) {
	t.Parallel()

	root := templateDef{SourcePath: "web/routes/root.templ"}
	layouts := map[string]templateDef{
		"": {RouteID: "", SourcePath: "web/routes/layout.templ"},
	}
	metas := []routeMeta{{
		RouteID:       "dashboard",
		PublicRouteID: "dashboard",
		Page:          templateDef{SourcePath: "web/routes/dashboard/page.templ"},
	}, {
		RouteID:       "details",
		PublicRouteID: "details",
		Page:          templateDef{SourcePath: "web/routes/details/page.templ"},
	}}
	slotOwners := map[string][]slotDef{
		"": {{
			Name:         "aside",
			RootInternal: "_slot__aside",
			Default:      &templateDef{RouteID: "_slot__aside", SourcePath: "web/routes/_slot__aside/default.templ"},
			Pages: []routeMeta{{
				RouteID:       "_slot__aside/details",
				PublicRouteID: "details",
				Page:          templateDef{SourcePath: "web/routes/_slot__aside/details/page.templ"},
			}},
			Layouts: map[string]templateDef{
				"_slot__aside": {RouteID: "_slot__aside", SourcePath: "web/routes/_slot__aside/layout.templ"},
			},
		}},
	}

	specs := clientAssetRouteSpecs(root, metas, layouts, slotOwners)
	require.Len(t, specs, 2)
	require.Equal(t, "dashboard", specs[0].RouteID)
	require.Equal(t, []string{
		"web/routes/root.templ",
		"web/routes/layout.templ",
		"web/routes/_slot__aside/layout.templ",
		"web/routes/_slot__aside/default.templ",
		"web/routes/_slot__aside/details/page.templ",
		"web/routes/dashboard/page.templ",
	}, specs[0].TemplatePaths)
	require.Equal(t, []clientassets.CSSBundleSpec{
		{
			OwnerTemplatePath: "web/routes/root.templ",
			TemplatePaths:     []string{"web/routes/root.templ"},
		},
		{
			OwnerTemplatePath: "web/routes/layout.templ",
			TemplatePaths: []string{
				"web/routes/layout.templ",
				"web/routes/_slot__aside/layout.templ",
				"web/routes/_slot__aside/default.templ",
				"web/routes/_slot__aside/details/page.templ",
			},
		},
		{
			OwnerTemplatePath: "web/routes/dashboard/page.templ",
			TemplatePaths:     []string{"web/routes/dashboard/page.templ"},
		},
	}, specs[0].CSSBundles)

	require.Equal(t, "details", specs[1].RouteID)
	require.Equal(t, []string{
		"web/routes/root.templ",
		"web/routes/layout.templ",
		"web/routes/_slot__aside/layout.templ",
		"web/routes/_slot__aside/default.templ",
		"web/routes/_slot__aside/details/page.templ",
		"web/routes/details/page.templ",
	}, specs[1].TemplatePaths)
	require.Equal(t, []clientassets.CSSBundleSpec{
		{
			OwnerTemplatePath: "web/routes/root.templ",
			TemplatePaths:     []string{"web/routes/root.templ"},
		},
		{
			OwnerTemplatePath: "web/routes/layout.templ",
			TemplatePaths: []string{
				"web/routes/layout.templ",
				"web/routes/_slot__aside/layout.templ",
				"web/routes/_slot__aside/default.templ",
				"web/routes/_slot__aside/details/page.templ",
			},
		},
		{
			OwnerTemplatePath: "web/routes/details/page.templ",
			TemplatePaths:     []string{"web/routes/details/page.templ"},
		},
	}, specs[1].CSSBundles)
}

func TestGenerateBundleSourceWiresStaticAssetHookWhenPresent(t *testing.T) {
	t.Parallel()

	source, err := generateBundleSource(projectlayout.ProjectLayout{
		AppModulePath:   testAppModulePath,
		GeneratedImport: "web/generated",
		ViewImport:      "web/view",
	}, viewcontract.Inspection{
		HasStaticAssetBasePathHook: true,
	})
	require.NoError(t, err)
	require.Contains(t, string(source), "OnStaticAssetBasePathResolved: view.SetStaticAssetBasePath,")
}

func TestRegistryGenerationUsesNotFoundResolverModels(t *testing.T) {
	t.Parallel()

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		[]routeMeta{{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.RootPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		}},
		nil,
		nil,
		nil,
		templateDef{Kind: rootTemplate, RouteID: "", ModuleName: "r_root_root"},
		map[string]templateDef{},
		map[string]templateDef{
			"": {Kind: notFoundTemplate, RouteID: "", ModuleName: "r_not_found_root", ModelType: "view.RootNotFoundView"},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "view, err := resolvers.ResolveRootNotFound")
	require.Contains(t, text, "component := r_not_found_root.NotFound(view, pathValue)")
	require.NotContains(t, text, "resolveNotFoundView")
	require.NotContains(t, text, "view.NewNotFoundView")
	require.NotContains(t, text, "appCtx.I18n(r)")
	require.NotContains(t, text, "NewErrorView")
}

func TestRewritePackageDeclarationAddsGeneratedMarker(t *testing.T) {
	source := "package appsrc\n\nimport (\n\t\"fmt\"\n)\n"

	rewritten, err := rewritePackageDeclaration([]byte(source), "r_page_root")
	require.NoError(t, err)

	text := string(rewritten)
	require.True(t, strings.HasPrefix(text, "package r_page_root\n"+generatedTemplHeader+"\n"))
	require.Equal(t, 1, strings.Count(text, generatedTemplHeader))
}

func TestRewritePackageDeclarationKeepsSingleGeneratedMarker(t *testing.T) {
	source := "package appsrc\n\n" + generatedTemplHeader + "\n\ntempl Page() { <div></div> }\n"

	rewritten, err := rewritePackageDeclaration([]byte(source), "r_page_root")
	require.NoError(t, err)

	text := string(rewritten)
	require.Equal(t, 1, strings.Count(text, generatedTemplHeader))
	require.True(t, strings.HasPrefix(text, "package r_page_root\n"))
}

func TestRunUsesExplicitProjectLayout(t *testing.T) {
	rootDir := t.TempDir()
	appDir := filepath.Join(rootDir, "internal", "web", "app")
	genDir := filepath.Join(rootDir, "internal", "web", "gen")
	resolverDir := filepath.Join(rootDir, "internal", "web", "resolvers")
	viewDir := filepath.Join(rootDir, "web", "view")

	writeMinimalZeroArgViewPackage(t, viewDir)

	writeTestFile(t, filepath.Join(appDir, "root.templ"), `package appsrc

import "github.com/RevoTale/no-js/framework/metagen"

templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) { @child }
`)
	writeTestFile(t, filepath.Join(appDir, "404.templ"), `package appsrc

import "example.com/app/web/view"

templ NotFound(model view.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "error.templ"), `package appsrc

import "example.com/app/web/view"

templ Error(model view.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div>notes</div> }
`)

	err := Run(Config{
		Layout: projectlayout.ProjectLayout{
			RootDir:         rootDir,
			RoutesDir:       appDir,
			GeneratedDir:    genDir,
			GeneratedImport: "web/generated",
			ResolversDir:    resolverDir,
			ViewDir:         viewDir,
			ViewImport:      "web/view",
			AppModulePath:   testAppModulePath,
		},
	})
	require.NoError(t, err)

	_, statErr := os.Stat(filepath.Join(genDir, "registry_gen.go"))
	require.NoError(t, statErr)
	_, statErr = os.Stat(filepath.Join(genDir, "server_gen.go"))
	require.Error(t, statErr)
	_, statErr = os.Stat(filepath.Join(resolverDir, generatedResolverFileName))
	require.NoError(t, statErr)
	_, statErr = os.Stat(filepath.Join(genDir, "r_error_root"))
	require.Error(t, statErr)

	registrySource, err := os.ReadFile(filepath.Join(genDir, "registry_gen.go"))
	require.NoError(t, err)
	require.NotContains(t, string(registrySource), "ErrorPage:")
}

func TestRunDoesNotRequireSystemPageConstructors(t *testing.T) {
	rootDir := t.TempDir()
	appDir := filepath.Join(rootDir, "web", "routes")
	genDir := filepath.Join(rootDir, "web", "generated")
	resolverDir := filepath.Join(rootDir, "web", "resolvers")
	viewDir := filepath.Join(rootDir, "web", "view")

	writeMinimalRouteTree(t, appDir)
	writeTypedSystemPageViewPackage(t, viewDir)

	err := Run(Config{
		Layout: projectlayout.ProjectLayout{
			RootDir:         rootDir,
			RoutesDir:       appDir,
			GeneratedDir:    genDir,
			GeneratedImport: "web/generated",
			ResolversDir:    resolverDir,
			ViewDir:         viewDir,
			ViewImport:      "web/view",
			AppModulePath:   testAppModulePath,
		},
	})
	require.NoError(t, err)

	registrySource, err := os.ReadFile(filepath.Join(genDir, "registry_gen.go"))
	require.NoError(t, err)
	require.NotContains(t, string(registrySource), "NewNotFoundView")
}
