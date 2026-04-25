package approutegen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/RevoTale/no-js/internal/bundler/clientassets"
	"github.com/RevoTale/no-js/internal/bundler/viewcontract"
	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

const testAppModulePath = "example.com/app"

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

func TestParsePageViewType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(
		t,
		pagePath,
		`package appsrc

import "example.com/app/web/view"

templ Page(model view.NotePageView) { <div/> }
`,
	)

	viewType, err := parsePageViewType(pagePath)
	require.NoError(t, err)
	require.Equal(t, "view.NotePageView", viewType)
}

func TestParsePageViewTypeRejectsNonViewType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(t, pagePath, "package appsrc\n\ntempl Page(view note.NotePageView) { <div/> }\n")

	_, err := parsePageViewType(pagePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "view-qualified")
}

func TestParseLayoutTemplateModelType(t *testing.T) {
	root := t.TempDir()
	rootValidPath := filepath.Join(root, "root_layout_valid.templ")
	rootInvalidPath := filepath.Join(root, "root_layout_invalid.templ")
	childValidPath := filepath.Join(root, "child_layout_valid.templ")
	childValidNamedPath := filepath.Join(root, "child_layout_valid_named.templ")
	childInvalidPath := filepath.Join(root, "child_layout_invalid.templ")
	writeTestFile(
		t,
		rootValidPath,
		`package appsrc

import (
  "github.com/RevoTale/no-js/framework/metagen"
  "example.com/app/web/view"
)

templ Layout(meta metagen.Metadata, model view.RootLayoutView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		rootInvalidPath,
		`package appsrc

import (
  "github.com/RevoTale/no-js/framework/metagen"
  "example.com/app/web/view"
)

templ Layout(meta metagen.Metadata, model runtime.RootLayoutView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childValidPath,
		`package appsrc

import "example.com/app/web/view"

templ Layout(model view.RootLayoutView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childValidNamedPath,
		`package appsrc

import "example.com/app/web/view"

templ Layout(layoutView view.RootLayoutView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childInvalidPath,
		`package appsrc

import (
  "github.com/RevoTale/no-js/framework/metagen"
  "example.com/app/web/view"
)

templ Layout(meta metagen.Metadata, model view.RootLayoutView, child templ.Component) { @child }
`,
	)

	rootModelType, err := parseLayoutTemplateModelType(templateDef{RouteID: "", SourcePath: rootValidPath}, nil)
	require.NoError(t, err)
	require.Equal(t, "view.RootLayoutView", rootModelType)
	_, err = parseLayoutTemplateModelType(templateDef{RouteID: "", SourcePath: rootInvalidPath}, nil)
	require.Error(t, err)
	childValidTemplate := templateDef{RouteID: "author/_param__slug", SourcePath: childValidPath}
	childModelType, err := parseLayoutTemplateModelType(childValidTemplate, nil)
	require.NoError(t, err)
	require.Equal(t, "view.RootLayoutView", childModelType)
	childValidNamedTemplate := templateDef{RouteID: "author/_param__slug", SourcePath: childValidNamedPath}
	childNamedModelType, err := parseLayoutTemplateModelType(childValidNamedTemplate, nil)
	require.NoError(t, err)
	require.Equal(t, "view.RootLayoutView", childNamedModelType)
	childInvalidTemplate := templateDef{RouteID: "author/_param__slug", SourcePath: childInvalidPath}
	_, err = parseLayoutTemplateModelType(childInvalidTemplate, nil)
	require.Error(t, err)
}

func TestParseNotFoundTemplateModelType(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "404_valid.templ")
	invalidPath := filepath.Join(root, "404_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "example.com/app/web/view"

templ NotFound(model view.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView, path string) { <div>{ path }</div> }
`,
	)

	modelType, err := parseNotFoundTemplateModelType(validPath)
	require.NoError(t, err)
	require.Equal(t, "view.RootLayoutView", modelType)
	_, err = parseNotFoundTemplateModelType(invalidPath)
	require.Error(t, err)
}

func TestValidateRootTemplateSignature(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "root_valid.templ")
	invalidPath := filepath.Join(root, "root_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "github.com/RevoTale/no-js/framework/metagen"

templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

templ RootLayout(locale string, child templ.Component) { @child }
`,
	)

	require.NoError(t, validateRootTemplateSignature(validPath))
	require.Error(t, validateRootTemplateSignature(invalidPath))
}

func TestValidateNoDocumentTagsAllowsHeader(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "layout_valid.templ")
	invalidPath := filepath.Join(root, "layout_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

templ Layout() {
	<header>ok</header>
}
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

templ Layout() {
	<head><title>bad</title></head>
}
`,
	)

	require.NoError(t, validateNoDocumentTags(validPath))
	require.Error(t, validateNoDocumentTags(invalidPath))
}

func TestBuildRouteMetasPageOnly(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	rootTemplate := `package appsrc

import "example.com/app/web/view"

templ Page(model view.NotesPageView) { <div id="notes-content"></div> }
`
	authorTemplate := `package appsrc

import "example.com/app/web/view"

templ Page(model view.AuthorPageView) { <div id="notes-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "page.templ"), rootTemplate)
	writeTestFile(t, filepath.Join(appRoot, "author", "_param__slug", "page.templ"), authorTemplate)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	metas, err := buildRouteMetas(routes.Pages, projectlayout.ProjectLayout{})
	require.NoError(t, err)

	byRoute := map[string]routeMeta{}
	for _, meta := range metas {
		byRoute[meta.RouteID] = meta
	}

	rootMeta, ok := byRoute[""]
	require.True(t, ok)
	require.Equal(t, "view.NotesPageView", rootMeta.PageViewType)

	authorMeta, ok := byRoute["author/_param__slug"]
	require.True(t, ok)
	require.Equal(t, "view.AuthorPageView", authorMeta.PageViewType)
}

func TestBuildRouteMetasAllowsNonPageViewSuffix(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	pageTemplate := `package appsrc

import "example.com/app/web/view"

templ Page(model view.NoteView) { <div id="note-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "note", "_param__slug", "page.templ"), pageTemplate)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	metas, err := buildRouteMetas(routes.Pages, projectlayout.ProjectLayout{})
	require.NoError(t, err)
	require.Len(t, metas, 1)
	require.Equal(t, "view.NoteView", metas[0].PageViewType)
}

func TestResolverNamespaceGenerationDeterministic(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
		},
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
		},
	}

	first, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		nil,
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.NoError(t, err)
	second, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		nil,
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.NoError(t, err)
	require.True(t, bytes.Equal(first, second))
	require.Contains(t, string(first), "var _ RouteResolver = (*Resolver)(nil)")
}

func TestRegistryGenerationUsesSingleResolverNamespace(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
			Page:           templateDef{ModuleName: "r_page_author_param_slug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "route_resolvers \"example.com/app/web/resolvers\"")
	require.NotContains(t, text, "rr_")
	require.Contains(t, text, "func NewRouteResolvers() RouteResolvers")
	require.Contains(t, text, "return &route_resolvers.Resolver{}")
	require.NotContains(t, text, "type NotFoundViewResolver interface")
	require.Contains(
		t,
		text,
		"func NotFoundPage(resolvers RouteResolvers) func(appCtx *view.Context, "+
			"r *http.Request, notFound framework.NotFoundContext) (templ.Component, error)",
	)
	require.Contains(t, text, "framework.PageOnlyRouteHandler")
	require.NotContains(t, text, "PageAndLiveRouteHandler")
	require.NotContains(t, text, "/.live/")
	require.NotContains(t, text, "ParseRootLiveState")
	require.Contains(t, text, "return renderNotFoundPage(resolvers, appCtx, r, notFound)")
	require.Contains(t, text, "view, err := resolvers.ResolveRootNotFound")
	require.Contains(t, text, "RootLayout: r_root_root.RootLayout")
	require.Contains(t, text, "MetaGenContextChain: []framework.PageMetaGenContext")
	require.NotContains(t, text, "ErrorPage:")
	require.NotContains(t, text, "r_error_root")
}

func TestRegistryGenerationEmitsClientAssets(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.RootPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}
	plan := clientassets.Plan{
		RouteAssets: map[string]metagen.ClientAssets{
			"": {
				Stylesheets:   []string{"routes/index.css"},
				ModuleScripts: []string{"routes/index.js"},
			},
		},
		NotFoundAssets: map[string]metagen.ClientAssets{
			"": {Stylesheets: []string{"routes/404.css"}},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{Kind: rootTemplate, RouteID: "", ModuleName: "r_root_root"},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
		plan,
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "ClientAssets: metagen.ClientAssets{")
	require.Contains(t, text, `"routes/index.css"`)
	require.Contains(t, text, `"routes/index.js"`)
	require.Contains(t, text, "metagen.MergeManagedClientAssets(requestContext(r), meta, notFoundClientAssets(routeID))")
	require.Contains(t, text, "func notFoundClientAssets(routeID string) metagen.ClientAssets")
	require.Contains(t, text, `"routes/404.css"`)
}

func TestRegistryGenerationRequiresRootNotFoundTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	_, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing root 404")
}

func TestRegistryGenerationCombinesRootAndDynamicNotFoundCases(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
			Page:           templateDef{ModuleName: "r_page_author_param_slug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
			"author/_param__slug": {
				Kind:       notFoundTemplate,
				RouteID:    "author/_param__slug",
				ModuleName: "r_not_found_author_param_slug",
				ModelType:  "view.AuthorNotFoundView",
			},
		},
	)
	require.NoError(t, err)
	require.Contains(t, string(registry), `case "", "author/_param__slug":`)
}

func TestRegistryGenerationDoesNotRequireRootErrorTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)
	require.NotContains(t, string(registry), "ErrorPage:")
}

func TestRegistryGenerationDoesNotWireErrorTemplateModules(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "author/_param__slug/note/_param__noteSlug",
			RouteName:      "AuthorParamSlugNoteParamNoteslug",
			ParamsTypeName: "AuthorParamSlugNoteParamNoteslugParams",
			Params: []routeParamDef{
				{Name: "slug", FieldName: "Slug"},
				{Name: "noteSlug", FieldName: "Noteslug"},
			},
			PageViewType: "view.NotePageView",
			Page:         templateDef{ModuleName: "r_page_author_param_slug_note_param_noteslug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.NotContains(t, text, "ErrorPage:")
	require.NotContains(t, text, "r_error_")
}

func TestGenerateSourcePackageParamsFileUsesPublicRouteID(t *testing.T) {
	source, err := generateSourcePackageParamsFile(sourcePackageDef{
		InternalRouteID: "_group__marketing/posts/_param__slug",
		PublicRouteID:   "posts/_param__slug",
		ParamsTypeName:  "GroupMarketingPostsParamSlugParams",
		Params: []routeParamDef{
			{Name: "slug", FieldName: "Slug", Type: "string"},
		},
		Package: "r_source_group_marketing_posts_param_slug",
	})
	require.NoError(t, err)
	require.Contains(t, string(source), `router.MatchPathPattern("/posts/_param__slug", requestPath)`)
}

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

func TestGenerateBundleSourceWiresTemplCSSRegistry(t *testing.T) {
	t.Parallel()

	source, err := generateBundleSource(projectlayout.ProjectLayout{
		AppModulePath:   testAppModulePath,
		GeneratedImport: "web/generated",
		ViewImport:      "web/view",
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

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}

func zeroArgViewInspection() viewcontract.Inspection {
	return viewcontract.Inspection{}
}

func writeMinimalZeroArgViewPackage(t *testing.T, viewDir string) {
	t.Helper()

	writeTestFile(t, filepath.Join(viewDir, "context.go"), `package view

import (
	"net/http"
	"net/url"
)

type Context struct{}

func (c *Context) ResolveRoot(*http.Request) *url.URL {
	root, _ := url.Parse("https://example.com")
	return root
}
`)
	writeTestFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return model.PageTitle
}

func NewNotFoundView() RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}
`)
}

func writeTypedSystemPageViewPackage(t *testing.T, viewDir string) {
	t.Helper()

	writeTestFile(t, filepath.Join(viewDir, "context.go"), `package view

import (
	"net/http"
	"net/url"
)

type Context struct{}
type Messages struct{}

func (c *Context) ResolveRoot(*http.Request) *url.URL {
	root, _ := url.Parse("https://example.com")
	return root
}

func (c *Context) I18n(*http.Request) *Messages {
	return nil
}
`)

	writeTestFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return model.PageTitle
}

func NewNotFoundView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}
`)
}

func writeMinimalRouteTree(t *testing.T, routesDir string) {
	t.Helper()

	writeTestFile(t, filepath.Join(routesDir, "root.templ"), `package appsrc

import "github.com/RevoTale/no-js/framework/metagen"

templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) { @child }
`)
	writeTestFile(t, filepath.Join(routesDir, "404.templ"), `package appsrc

import "example.com/app/web/view"

templ NotFound(model view.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(routesDir, "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(model view.RootPageView) { <div>home</div> }
`)
}
