package approutegen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	writeTestFile(t, filepath.Join(appRoot, "author", "[slug]", "page.templ"), "package appsrc\n")

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)
	require.Len(t, routes.Pages, 2)
	require.Equal(t, "author/[slug]", routes.Pages[0].RouteID)
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
	require.Contains(t, err.Error(), "use [param]")
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

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "[slug]", "404.templ"),
		`package appsrc

import "example.com/app/web/view"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "[slug]", "page.templ"),
		`package appsrc

import "example.com/app/web/view"

templ Page(view runtime.AuthorPageView) { <div id="notes-content"></div> }
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	_, ok := routes.NotFounds[""]
	require.True(t, ok)
	_, ok = routes.NotFounds["author/[slug]"]
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

	view "example.com/app/web/view"
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

	view "example.com/app/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	return nil, nil
}

func GenerateSitemaps(runtime framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapID, error) {
	return nil, nil
}

func SitemapByID(
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

	view "example.com/app/web/view"
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
	require.Equal(t, filepath.Join(appRoot, "sitemap.go"), routes.Discovery.SitemapFile)
	require.Equal(t, filepath.Join(appRoot, "feed.go"), routes.Discovery.FeedFile)

	err = validateDiscoveryConventions(projectlayout.ProjectLayout{
		AppModulePath: testAppModulePath,
		RoutesImport:  "web/routes",
		ViewImport:    "web/view",
	}, &routes.Discovery)
	require.NoError(t, err)
	require.True(t, routes.Discovery.HasRobots)
	require.True(t, routes.Discovery.HasSitemap)
	require.True(t, routes.Discovery.HasGenerateSitemaps)
	require.True(t, routes.Discovery.HasSitemapByID)
	require.True(t, routes.Discovery.HasFeed)
}

func TestValidateDiscoveryConventionsRejectsGenerateWithoutSitemapByID(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(
		t,
		filepath.Join(appRoot, "sitemap.go"),
		`package routes

import (
	"net/http"

	view "example.com/app/web/view"
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
	require.Contains(t, err.Error(), "GenerateSitemaps requires SitemapByID")
}

func TestParsePageViewType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(
		t,
		pagePath,
		`package appsrc

import "example.com/app/web/view"

templ Page(view runtime.NotePageView) { <div/> }
`,
	)

	viewType, err := parsePageViewType(pagePath)
	require.NoError(t, err)
	require.Equal(t, "runtime.NotePageView", viewType)
}

func TestParsePageViewTypeRejectsNonRuntimeType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(t, pagePath, "package appsrc\n\ntempl Page(view note.NotePageView) { <div/> }\n")

	_, err := parsePageViewType(pagePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "runtime-qualified")
}

func TestValidateLayoutTemplateSignature(t *testing.T) {
	root := t.TempDir()
	rootValidPath := filepath.Join(root, "root_layout_valid.templ")
	rootInvalidPath := filepath.Join(root, "root_layout_invalid.templ")
	childValidPath := filepath.Join(root, "child_layout_valid.templ")
	childInvalidPath := filepath.Join(root, "child_layout_invalid.templ")
	writeTestFile(
		t,
		rootValidPath,
		`package appsrc

import (
  "github.com/RevoTale/no-js/framework/metagen"
  "example.com/app/web/view"
)

templ Layout(meta metagen.Metadata, view runtime.RootLayoutView, child templ.Component) { @child }
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

templ Layout(meta metagen.Metadata, view runtime.NotesPageView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childValidPath,
		`package appsrc

import "example.com/app/web/view"

templ Layout(view runtime.RootLayoutView, child templ.Component) { @child }
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

templ Layout(meta metagen.Metadata, view runtime.RootLayoutView, child templ.Component) { @child }
`,
	)

	require.NoError(t, validateLayoutTemplateSignature(templateDef{RouteID: "", SourcePath: rootValidPath}))
	require.Error(t, validateLayoutTemplateSignature(templateDef{RouteID: "", SourcePath: rootInvalidPath}))
	childValidTemplate := templateDef{RouteID: "author/[slug]", SourcePath: childValidPath}
	require.NoError(t, validateLayoutTemplateSignature(childValidTemplate))
	childInvalidTemplate := templateDef{RouteID: "author/[slug]", SourcePath: childInvalidPath}
	require.Error(t, validateLayoutTemplateSignature(childInvalidTemplate))
}

func TestValidateNotFoundTemplateSignature(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "404_valid.templ")
	invalidPath := filepath.Join(root, "404_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "example.com/app/web/view"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

import "example.com/app/web/view"

templ Page(view runtime.NotesPageView, path string) { <div>{ path }</div> }
`,
	)

	require.NoError(t, validateNotFoundTemplateSignature(validPath))
	require.Error(t, validateNotFoundTemplateSignature(invalidPath))
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

func TestValidateErrorTemplateSignature(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "error_valid.templ")
	invalidPath := filepath.Join(root, "error_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "example.com/app/web/view"

templ Error(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

import "example.com/app/web/view"

templ Error(view runtime.NotePageView, path string) { <div>{ path }</div> }
`,
	)

	require.NoError(t, validateErrorTemplateSignature(validPath))
	require.Error(t, validateErrorTemplateSignature(invalidPath))
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

templ Page(view runtime.NotesPageView) { <div id="notes-content"></div> }
`
	authorTemplate := `package appsrc

import "example.com/app/web/view"

templ Page(view runtime.AuthorPageView) { <div id="notes-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "page.templ"), rootTemplate)
	writeTestFile(t, filepath.Join(appRoot, "author", "[slug]", "page.templ"), authorTemplate)

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
	require.Equal(t, "runtime.NotesPageView", rootMeta.PageViewType)

	authorMeta, ok := byRoute["author/[slug]"]
	require.True(t, ok)
	require.Equal(t, "runtime.AuthorPageView", authorMeta.PageViewType)
}

func TestBuildRouteMetasAllowsNonPageViewSuffix(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	pageTemplate := `package appsrc

import "example.com/app/web/view"

templ Page(view runtime.NoteView) { <div id="note-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "note", "[slug]", "page.templ"), pageTemplate)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	require.NoError(t, err)

	metas, err := buildRouteMetas(routes.Pages, projectlayout.ProjectLayout{})
	require.NoError(t, err)
	require.Len(t, metas, 1)
	require.Equal(t, "runtime.NoteView", metas[0].PageViewType)
}

func TestResolverNamespaceGenerationDeterministic(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "runtime.NotesPageView",
		},
		{
			RouteID:        "author/[slug]",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "runtime.AuthorPageView",
		},
	}

	first, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		map[string]templateDef{},
	)
	require.NoError(t, err)
	second, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
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
			PageViewType:   "runtime.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
		{
			RouteID:        "author/[slug]",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "runtime.AuthorPageView",
			Page:           templateDef{ModuleName: "r_page_author_param_slug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		metas,
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
			},
		},
		map[string]templateDef{
			"": {
				Kind:       errorTemplate,
				RouteID:    "",
				ModuleName: "r_error_root",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "route_resolvers \"example.com/app/web/resolvers\"")
	require.NotContains(t, text, "rr_")
	require.Contains(t, text, "func NewRouteResolvers() RouteResolvers")
	require.Contains(t, text, "return &route_resolvers.Resolver{}")
	require.Contains(t, text, "framework.PageOnlyRouteHandler")
	require.NotContains(t, text, "PageAndLiveRouteHandler")
	require.NotContains(t, text, "/.live/")
	require.NotContains(t, text, "ParseRootLiveState")
	require.Contains(t, text, "func NotFoundPage(notFound framework.NotFoundContext) templ.Component")
	require.Contains(t, text, "RootLayout: r_root_root.RootLayout")
	require.Contains(t, text, "MetaGenChain: []framework.PageMetaGen")
	require.Contains(t, text, "ErrorPage: func(locale string, path string) templ.Component")
}

func TestRegistryGenerationRequiresRootNotFoundTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "runtime.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	_, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		metas,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       errorTemplate,
				RouteID:    "",
				ModuleName: "r_error_root",
			},
		},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing root 404")
}

func TestRegistryGenerationRequiresRootErrorTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "runtime.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	_, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		metas,
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
			},
		},
		map[string]templateDef{},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing root error")
}

func TestRegistryGenerationWiresNearestErrorTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "author/[slug]/note/[noteSlug]",
			RouteName:      "AuthorParamSlugNoteParamNoteslug",
			ParamsTypeName: "AuthorParamSlugNoteParamNoteslugParams",
			Params: []routeParamDef{
				{Name: "slug", FieldName: "Slug"},
				{Name: "noteSlug", FieldName: "Noteslug"},
			},
			PageViewType: "runtime.NotePageView",
			Page:         templateDef{ModuleName: "r_page_author_param_slug_note_param_noteslug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		metas,
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
			},
		},
		map[string]templateDef{
			"": {
				Kind:       errorTemplate,
				RouteID:    "",
				ModuleName: "r_error_root",
			},
			"author/[slug]": {
				Kind:       errorTemplate,
				RouteID:    "author/[slug]",
				ModuleName: "r_error_author_param_slug",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "component := r_error_author_param_slug.Error(view, pathValue)")
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

	writeTestFile(t, filepath.Join(appDir, "root.templ"), `package appsrc

import "github.com/RevoTale/no-js/framework/metagen"

templ RootLayout(meta metagen.Metadata, locale string, child templ.Component) { @child }
`)
	writeTestFile(t, filepath.Join(appDir, "404.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "error.templ"), `package appsrc

import "example.com/app/web/view"

templ Error(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "page.templ"), `package appsrc

import "example.com/app/web/view"

templ Page(view runtime.NotesPageView) { <div>notes</div> }
`)

	err := Run(Config{
		Layout: projectlayout.ProjectLayout{
			RootDir:         rootDir,
			RoutesDir:       appDir,
			GeneratedDir:    genDir,
			GeneratedImport: "web/generated",
			ResolversDir:    resolverDir,
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
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
