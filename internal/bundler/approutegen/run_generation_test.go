package approutegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestRunGeneratesMethodRoutesFromGroupedRouteGo(t *testing.T) {
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

templ Error(model view.DoesNotExist, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "_group__marketing", "posts", "_param__slug", "route.go"), `package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params GroupMarketingPostsParamSlugParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
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

	registrySource, err := os.ReadFile(filepath.Join(genDir, "registry_gen.go"))
	require.NoError(t, err)
	registryText := string(registrySource)
	require.Contains(t, registryText, "framework.MethodOnlyRouteHandler")
	require.Contains(t, registryText, `"_group__marketing/posts/_param__slug"`)
	require.Contains(t, registryText, `"/posts/_param__slug"`)

	paramsSource, err := os.ReadFile(filepath.Join(genDir, "r_source_group_marketing_posts_param_slug", "params_gen.go"))
	require.NoError(t, err)
	require.Contains(t, string(paramsSource), `router.MatchPathPattern("/posts/_param__slug", requestPath)`)
}

func TestRunGeneratesSlotComposeWithoutSlotMetadata(t *testing.T) {
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
	writeTestFile(t, filepath.Join(appDir, "dashboard", "layout.templ"), `package appsrc

import "example.com/app/web/view"

templ Layout(model view.RootLayoutView, child templ.Component, analytics templ.Component) {
	@child
	@analytics
}
`)
	writeTestFile(t, filepath.Join(appDir, "dashboard", "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div>dashboard</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "dashboard", "_slot__analytics", "default.templ"), `package appsrc

import "example.com/app/web/view"

templ Default(model view.RootLayoutView) { <div>default analytics</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "dashboard", "_slot__analytics", "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div>analytics</div> }
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

	registrySource, err := os.ReadFile(filepath.Join(genDir, "registry_gen.go"))
	require.NoError(t, err)
	registryText := string(registrySource)
	require.Contains(t, registryText, "func resolveDashboardAnalyticsSlot(")
	require.Contains(t, registryText, "func composeDashboardPage(")
	require.Contains(t, registryText, "var slotWG sync.WaitGroup")
	require.Contains(t, registryText, "go func() {")
	require.Contains(t, registryText, "component, err := resolveDashboardAnalyticsSlot")
	require.Contains(t, registryText, "dashboardAnalyticsSlot = component")
	require.Contains(t, registryText, "resolvers.ResolveDashboardLayout")
	require.Contains(t, registryText, "r_layout_dashboard.Layout(dashboardLayoutView, component, dashboardAnalyticsSlot)")
	require.Contains(
		t,
		registryText,
		"component := r_default_dashboard_slot_analytics.Default(dashboardslotanalyticsDefaultView)",
	)

	resolverSource, err := os.ReadFile(filepath.Join(resolverDir, generatedResolverFileName))
	require.NoError(t, err)
	require.NotContains(t, string(resolverSource), "MetaGenDashboardSlotAnalyticsPage")
}

func TestRunGeneratesGroupedPageSlotAndMethodRouteTree(t *testing.T) {
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
	writeTestFile(t, filepath.Join(appDir, "_group__marketing", "dashboard", "layout.templ"), `package appsrc

import "example.com/app/web/view"

templ Layout(model view.RootLayoutView, child templ.Component, analytics templ.Component) {
	@child
	@analytics
}
`)
	writeTestFile(t, filepath.Join(appDir, "_group__marketing", "dashboard", "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div>dashboard</div> }
`)
	writeTestFile(
		t,
		filepath.Join(appDir, "_group__marketing", "dashboard", "_slot__analytics", "default.templ"),
		`package appsrc

import "example.com/app/web/view"

templ Default(model view.RootLayoutView) { <div>default analytics</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appDir, "_group__marketing", "dashboard", "_slot__analytics", "page.templ"),
		`package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div>analytics</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appDir, "_group__marketing", "dashboard", "export", "route.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params GroupMarketingDashboardExportParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
`,
	)

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
	registryText := string(registrySource)
	require.Contains(t, registryText, `"_group__marketing/dashboard"`)
	require.Contains(t, registryText, `"/dashboard"`)
	require.Contains(t, registryText, "func resolveGroupMarketingDashboardAnalyticsSlot(")
	require.Contains(t, registryText, "func composeGroupMarketingDashboardPage(")
	require.Contains(t, registryText, "var slotWG sync.WaitGroup")
	require.Contains(t, registryText, "go func() {")
	require.Contains(
		t,
		registryText,
		"component, err := resolveGroupMarketingDashboardAnalyticsSlot",
	)
	require.Contains(t, registryText, "groupmarketingdashboardAnalyticsSlot = component")
	require.Contains(
		t,
		registryText,
		"r_layout_group_marketing_dashboard.Layout("+
			"groupmarketingdashboardLayoutView, component, groupmarketingdashboardAnalyticsSlot)",
	)
	require.Contains(t, registryText, "framework.MethodOnlyRouteHandler")
	require.Contains(t, registryText, `"_group__marketing/dashboard/export"`)
	require.Contains(t, registryText, `"/dashboard/export"`)

	resolverSource, err := os.ReadFile(filepath.Join(resolverDir, generatedResolverFileName))
	require.NoError(t, err)
	require.NotContains(t, string(resolverSource), "MetaGenGroupMarketingDashboardSlotAnalyticsPage")
}

func TestGenerateDiscoverySourceWithoutDiscoveryRoutesStillDefinesExactHandlers(t *testing.T) {
	t.Parallel()

	source, err := generateDiscoverySource(projectlayout.ProjectLayout{
		AppModulePath: testAppModulePath,
		ViewImport:    "web/view",
	}, discoveryConventions{})
	require.NoError(t, err)

	text := string(source)
	require.Contains(t, text, "func DiscoveryBundle() *discovery.Bundle[*view.Context]")
	require.Contains(t, text, "func DiscoveryExactHandlers() []framework.RouteHandler[*view.Context]")
	require.Contains(t, text, "return nil")
}
