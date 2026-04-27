package approutegen

import (
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestDiscoverRouteFilesStaticAndDynamic(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "layout.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "notes", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "author", "_param__slug", "page.templ"), "package appsrc\n")

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Len(t, routes.Pages, 2)
	require.Equal(t, "author/_param__slug", routes.Pages[0].RouteID)
	require.Equal(t, "notes", routes.Pages[1].RouteID)
}

func TestDiscoverRouteFilesRejectsRouteLocalComponents(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "notes", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "notes", "components", "card.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "web/components")
}

func TestDiscoverRouteFilesRejectsRootComponentsDir(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "components", "note_card.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "web/components")
}

func TestDiscoverRouteFilesRejectsLegacyWildcardSyntax(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "note", "_slug", "page.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid reserved route segment")
}

func TestDiscoverRouteFilesSupportsReservedNamespace(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "author", "_param__slug", "page.templ"), "package appsrc\n")

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Len(t, routes.Pages, 1)
	require.Equal(t, "author/_param__slug", routes.Pages[0].RouteID)
	require.Equal(t, "author/_param__slug", routes.Pages[0].PublicRouteID)
}

func TestDiscoverRouteFilesRejectsUnknownReservedNamespace(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "_unknown__marketing", "page.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown reserved route segment")
}

func TestDiscoverRouteFilesRejectsPageAndRouteGoAtSameRoute(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "note", "_param__slug", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "note", "_param__slug", "route.go"), `package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
)

func GET(
	runtime framework.RuntimeContext[*view.Context],
	w http.ResponseWriter,
	r *http.Request,
	params NoteParamSlugParams,
) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
`)

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "route pattern conflict")
}

func TestDiscoverRouteFilesSupportsGroupedRoutes(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "_group__marketing", "about", "page.templ"), "package appsrc\n")

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Len(t, routes.Pages, 1)
	require.Equal(t, "_group__marketing/about", routes.Pages[0].RouteID)
	require.Equal(t, "about", routes.Pages[0].PublicRouteID)
}

func TestDiscoverRouteFilesSupportsSlots(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "dashboard", "layout.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "dashboard", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "dashboard", "_slot__analytics", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "dashboard", "_slot__analytics", "default.templ"), "package appsrc\n")

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Len(t, routes.Pages, 1)
	require.Len(t, routes.SlotPages, 1)
	require.Equal(t, "dashboard", routes.Pages[0].RouteID)
	require.Equal(t, "dashboard/_slot__analytics", routes.SlotPages[0].RouteID)
	require.Equal(t, "dashboard", routes.SlotPages[0].PublicRouteID)
	require.Equal(t, "dashboard", routes.SlotPages[0].SlotOwnerRouteID)
	require.Contains(t, routes.Defaults, "dashboard/_slot__analytics")
	require.Equal(t, []string{"analytics"}, routes.LayoutSlots["dashboard"])
}

func TestDiscoverRouteFilesRejectsRouteGoInsideSlot(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "dashboard", "layout.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "dashboard", "_slot__analytics", "route.go"), "package routes\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "route.go is not allowed inside slot directories")
}

func TestDiscoverRouteFilesRejectsDefaultOutsideSlotRoot(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "dashboard", "layout.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "dashboard", "_slot__analytics", "page.templ"), "package appsrc\n")
	writeTestFile(
		t,
		filepath.Join(appRoot, "dashboard", "_slot__analytics", "nested", "default.templ"),
		"package appsrc\n",
	)

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "default.templ is only allowed at the slot root")
}

func TestDiscoverRouteFilesRejectsSlotWithoutOwnerLayout(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "dashboard", "_slot__analytics", "page.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	require.Error(t, err)
	require.Contains(t, err.Error(), "slot owner \"dashboard\" requires a same-level layout.templ")
}

func TestBuildSlotOwnersIncludesDefaultOnlySlots(t *testing.T) {
	slotRoot := "_group__marketing/dashboard/_slot__analytics"

	slotOwners := buildSlotOwners(nil, nil, map[string]templateDef{
		slotRoot: {
			Kind:               defaultTemplate,
			RouteID:            slotRoot,
			InternalRouteID:    slotRoot,
			SlotOwnerRouteID:   "_group__marketing/dashboard",
			SlotRootInternalID: slotRoot,
			SlotName:           "analytics",
		},
	})

	slots, ok := slotOwners["_group__marketing/dashboard"]
	require.True(t, ok)
	require.Len(t, slots, 1)
	require.Equal(t, "analytics", slots[0].Name)
	require.Equal(t, slotRoot, slots[0].RootInternal)
	require.NotNil(t, slots[0].Default)
	require.Equal(t, slotRoot, slots[0].Default.RouteID)
}

func TestDiscoverRouteFilesCollectsNotFoundTemplates(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(
		t,
		filepath.Join(appRoot, "404.templ"),
		`package appsrc

import "example.com/app/web/view"

templ NotFound(model view.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__slug", "404.templ"),
		`package appsrc

import "example.com/app/web/view"

templ NotFound(model view.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__slug", "page.templ"),
		`package appsrc

import "example.com/app/web/view"

templ Page(model view.AuthorPageView) { <div id="notes-content"></div> }
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	_, ok := routes.NotFounds[""]
	require.True(t, ok)
	_, ok = routes.NotFounds["author/_param__slug"]
	require.True(t, ok)
}

func TestDiscoverRouteFilesCollectsDiscoveryConventions(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "root.templ"), "package appsrc\n")
	writeTestFile(
		t,
		filepath.Join(appRoot, "robots.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Robots(runtime framework.RuntimeContext[*view.Context], r *http.Request) (discovery.Robots, error) {
	return discovery.Robots{}, nil
}
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "sitemap.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	return nil, nil
}

func GenerateSitemaps(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapID, error) {
	return nil, nil
}


func SitemapChunk(
	runtime framework.RuntimeContext[*view.Context],
	r *http.Request,
	id string,
) ([]discovery.SitemapEntry, error) {
	return nil, nil
}
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "feed.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

	func Feed(runtime framework.RuntimeContext[*view.Context], r *http.Request) (discovery.FeedDocument, error) {
		return discovery.FeedDocument{}, nil
	}
	`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__slug", "sitemap.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	return nil, nil
}
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__slug", "feed.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtime framework.RuntimeContext[*view.Context], r *http.Request) (discovery.FeedDocument, error) {
	return discovery.FeedDocument{}, nil
}
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(appRoot, "robots.go"), routes.Discovery.RobotsFile)
	require.Len(t, routes.Discovery.Sitemaps, 2)
	require.Len(t, routes.Discovery.Feeds, 2)

	err = validateDiscoveryConventions(projectlayout.ProjectLayout{
		AppModulePath: testAppModulePath,
		RoutesImport:  "web/routes",
		ViewImport:    "web/view",
	}, &routes.Discovery)
	require.NoError(t, err)
	require.True(t, routes.Discovery.HasRobots)
	require.Equal(t, "author/_param__slug", routes.Discovery.Sitemaps[0].RouteID)
	require.True(t, routes.Discovery.Sitemaps[0].HasSitemap)
	require.False(t, routes.Discovery.Sitemaps[0].HasGenerateSitemaps)
	require.False(t, routes.Discovery.Sitemaps[0].HasSitemapChunk)
	require.Equal(t, "", routes.Discovery.Sitemaps[1].RouteID)
	require.True(t, routes.Discovery.Sitemaps[1].HasSitemap)
	require.True(t, routes.Discovery.Sitemaps[1].HasGenerateSitemaps)
	require.True(t, routes.Discovery.Sitemaps[1].HasSitemapChunk)
	require.Equal(t, "author/_param__slug", routes.Discovery.Feeds[0].RouteID)
	require.Equal(t, "", routes.Discovery.Feeds[1].RouteID)
}

func TestValidateDiscoveryConventionsRejectsIncompleteDynamicSitemap(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(
		t,
		filepath.Join(appRoot, "sitemap.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	return nil, nil
}

func GenerateSitemaps(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapID, error) {
	return nil, nil
}
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	err = validateDiscoveryConventions(projectlayout.ProjectLayout{
		AppModulePath: testAppModulePath,
		RoutesImport:  "web/routes",
		ViewImport:    "web/view",
	}, &routes.Discovery)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dynamic sitemaps require GenerateSitemaps and SitemapChunk")
}

func TestValidateDiscoveryConventionsRejectsNestedPatternConflicts(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__slug", "feed.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtime framework.RuntimeContext[*view.Context], r *http.Request) (discovery.FeedDocument, error) {
	return discovery.FeedDocument{}, nil
}
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "_param__id", "feed.go"),
		`package routes

import (
	"net/http"

	"example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtime framework.RuntimeContext[*view.Context], r *http.Request) (discovery.FeedDocument, error) {
	return discovery.FeedDocument{}, nil
}
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	err = validateDiscoveryConventions(projectlayout.ProjectLayout{
		AppModulePath: testAppModulePath,
		RoutesImport:  "web/routes",
		ViewImport:    "web/view",
	}, &routes.Discovery)
	require.Error(t, err)
	require.Contains(t, err.Error(), "discovery route pattern conflict")
}
