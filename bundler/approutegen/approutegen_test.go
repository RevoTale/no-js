package approutegen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RevoTale/no-js/bundler"
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
	if err != nil {
		t.Fatalf("discover routes: %v", err)
	}

	if len(routes.Pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(routes.Pages))
	}
	if routes.Pages[0].RouteID != "author/[slug]" {
		t.Fatalf("expected first route author/[slug], got %q", routes.Pages[0].RouteID)
	}
	if routes.Pages[1].RouteID != "notes" {
		t.Fatalf("expected second route notes, got %q", routes.Pages[1].RouteID)
	}
}

func TestDiscoverRouteFilesRejectsRouteLocalComponents(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "notes", "page.templ"), "package appsrc\n")
	writeTestFile(t, filepath.Join(appRoot, "notes", "components", "card.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	if err == nil {
		t.Fatal("expected route-local components error")
	}
	if !strings.Contains(err.Error(), "internal/web/components") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiscoverRouteFilesRejectsRootComponentsDir(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "components", "note_card.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	if err == nil {
		t.Fatal("expected root components error")
	}
	if !strings.Contains(err.Error(), "internal/web/components") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiscoverRouteFilesRejectsLegacyWildcardSyntax(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(t, filepath.Join(appRoot, "note", "_slug", "page.templ"), "package appsrc\n")

	_, err := discoverRouteFiles(appRoot, genRoot)
	if err == nil {
		t.Fatal("expected legacy wildcard syntax error")
	}
	if !strings.Contains(err.Error(), "use [param]") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDiscoverRouteFilesCollectsNotFoundTemplates(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	writeTestFile(
		t,
		filepath.Join(appRoot, "404.templ"),
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "[slug]", "404.templ"),
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		filepath.Join(appRoot, "author", "[slug]", "page.templ"),
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.AuthorPageView) { <div id="notes-content"></div> }
`,
	)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	if err != nil {
		t.Fatalf("discover routes: %v", err)
	}

	if _, ok := routes.NotFounds[""]; !ok {
		t.Fatalf("expected root 404 template")
	}
	if _, ok := routes.NotFounds["author/[slug]"]; !ok {
		t.Fatalf("expected nested author 404 template")
	}
}

func TestParsePageViewType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(
		t,
		pagePath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.NotePageView) { <div/> }
`,
	)

	viewType, err := parsePageViewType(pagePath)
	if err != nil {
		t.Fatalf("parse page view type: %v", err)
	}
	if viewType != "runtime.NotePageView" {
		t.Fatalf("expected runtime.NotePageView, got %q", viewType)
	}
}

func TestParsePageViewTypeRejectsNonRuntimeType(t *testing.T) {
	root := t.TempDir()
	pagePath := filepath.Join(root, "page.templ")
	writeTestFile(t, pagePath, "package appsrc\n\ntempl Page(view note.NotePageView) { <div/> }\n")

	_, err := parsePageViewType(pagePath)
	if err == nil {
		t.Fatal("expected runtime-qualified type error")
	}
	if !strings.Contains(err.Error(), "runtime-qualified") {
		t.Fatalf("unexpected error: %v", err)
	}
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
  "example.com/app/internal/web/runtime"
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
  "example.com/app/internal/web/runtime"
)

templ Layout(meta metagen.Metadata, view runtime.NotesPageView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childValidPath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Layout(view runtime.RootLayoutView, child templ.Component) { @child }
`,
	)
	writeTestFile(
		t,
		childInvalidPath,
		`package appsrc

import (
  "github.com/RevoTale/no-js/framework/metagen"
  "example.com/app/internal/web/runtime"
)

templ Layout(meta metagen.Metadata, view runtime.RootLayoutView, child templ.Component) { @child }
`,
	)

	if err := validateLayoutTemplateSignature(templateDef{RouteID: "", SourcePath: rootValidPath}); err != nil {
		t.Fatalf("expected valid root layout signature, got %v", err)
	}
	if err := validateLayoutTemplateSignature(templateDef{RouteID: "", SourcePath: rootInvalidPath}); err == nil {
		t.Fatal("expected invalid root layout signature error")
	}
	childValidTemplate := templateDef{RouteID: "author/[slug]", SourcePath: childValidPath}
	if err := validateLayoutTemplateSignature(childValidTemplate); err != nil {
		t.Fatalf("expected valid child layout signature, got %v", err)
	}
	childInvalidTemplate := templateDef{RouteID: "author/[slug]", SourcePath: childInvalidPath}
	if err := validateLayoutTemplateSignature(childInvalidTemplate); err == nil {
		t.Fatal("expected invalid child layout signature error")
	}
}

func TestValidateNotFoundTemplateSignature(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "404_valid.templ")
	invalidPath := filepath.Join(root, "404_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.NotesPageView, path string) { <div>{ path }</div> }
`,
	)

	if err := validateNotFoundTemplateSignature(validPath); err != nil {
		t.Fatalf("expected valid 404 signature, got %v", err)
	}
	if err := validateNotFoundTemplateSignature(invalidPath); err == nil {
		t.Fatal("expected invalid 404 signature error")
	}
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

	if err := validateRootTemplateSignature(validPath); err != nil {
		t.Fatalf("expected valid root signature, got %v", err)
	}
	if err := validateRootTemplateSignature(invalidPath); err == nil {
		t.Fatal("expected invalid root signature error")
	}
}

func TestValidateErrorTemplateSignature(t *testing.T) {
	root := t.TempDir()
	validPath := filepath.Join(root, "error_valid.templ")
	invalidPath := filepath.Join(root, "error_invalid.templ")
	writeTestFile(
		t,
		validPath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Error(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`,
	)
	writeTestFile(
		t,
		invalidPath,
		`package appsrc

import "example.com/app/internal/web/runtime"

templ Error(view runtime.NotePageView, path string) { <div>{ path }</div> }
`,
	)

	if err := validateErrorTemplateSignature(validPath); err != nil {
		t.Fatalf("expected valid error signature, got %v", err)
	}
	if err := validateErrorTemplateSignature(invalidPath); err == nil {
		t.Fatal("expected invalid error signature")
	}
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

	if err := validateNoDocumentTags(validPath); err != nil {
		t.Fatalf("header tag should be allowed, got %v", err)
	}
	if err := validateNoDocumentTags(invalidPath); err == nil {
		t.Fatal("expected head tag rejection")
	}
}

func TestBuildRouteMetasPageOnly(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	rootTemplate := `package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.NotesPageView) { <div id="notes-content"></div> }
`
	authorTemplate := `package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.AuthorPageView) { <div id="notes-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "page.templ"), rootTemplate)
	writeTestFile(t, filepath.Join(appRoot, "author", "[slug]", "page.templ"), authorTemplate)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	if err != nil {
		t.Fatalf("discover routes: %v", err)
	}

	metas, err := buildRouteMetas(routes.Pages, bundler.ProjectLayout{})
	if err != nil {
		t.Fatalf("build route metas: %v", err)
	}

	byRoute := map[string]routeMeta{}
	for _, meta := range metas {
		byRoute[meta.RouteID] = meta
	}

	rootMeta, ok := byRoute[""]
	if !ok {
		t.Fatalf("missing root route meta: %#v", byRoute)
	}
	if rootMeta.PageViewType != "runtime.NotesPageView" {
		t.Fatalf("expected root page view type, got %q", rootMeta.PageViewType)
	}

	authorMeta, ok := byRoute["author/[slug]"]
	if !ok {
		t.Fatalf("missing author route meta: %#v", byRoute)
	}
	if authorMeta.PageViewType != "runtime.AuthorPageView" {
		t.Fatalf("expected author page view type, got %q", authorMeta.PageViewType)
	}
}

func TestBuildRouteMetasAllowsNonPageViewSuffix(t *testing.T) {
	root := t.TempDir()
	appRoot := filepath.Join(root, "app")
	genRoot := filepath.Join(root, "gen")

	pageTemplate := `package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.NoteView) { <div id="note-content"></div> }
`
	writeTestFile(t, filepath.Join(appRoot, "note", "[slug]", "page.templ"), pageTemplate)

	routes, err := discoverRouteFiles(appRoot, genRoot)
	if err != nil {
		t.Fatalf("discover routes: %v", err)
	}

	metas, err := buildRouteMetas(routes.Pages, bundler.ProjectLayout{})
	if err != nil {
		t.Fatalf("build route metas: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("expected 1 route meta, got %d", len(metas))
	}
	if metas[0].PageViewType != "runtime.NoteView" {
		t.Fatalf("expected runtime.NoteView, got %q", metas[0].PageViewType)
	}
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
		bundler.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		map[string]templateDef{},
	)
	if err != nil {
		t.Fatalf("first generation failed: %v", err)
	}
	second, err := generateResolverNamespaceSource(
		bundler.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		map[string]templateDef{},
	)
	if err != nil {
		t.Fatalf("second generation failed: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("resolver namespace generation is not deterministic")
	}
	if !bytes.Contains(first, []byte("var _ RouteResolver = (*Resolver)(nil)")) {
		t.Fatalf("expected compile-time assertion in generated resolver namespace:\n%s", string(first))
	}
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
		bundler.ProjectLayout{GeneratedImport: "internal/web/gen", AppModulePath: testAppModulePath},
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
	if err != nil {
		t.Fatalf("generate registry: %v", err)
	}

	text := string(registry)
	if !strings.Contains(text, "route_resolvers \"example.com/app/internal/web/resolvers\"") {
		t.Fatalf("expected unified resolver namespace import in registry:\n%s", text)
	}
	if strings.Contains(text, "rr_") {
		t.Fatalf("did not expect per-route resolver aliases in registry:\n%s", text)
	}
	if !strings.Contains(text, "func NewRouteResolvers() RouteResolvers") {
		t.Fatalf("expected NewRouteResolvers constructor in registry:\n%s", text)
	}
	if !strings.Contains(text, "return &route_resolvers.Resolver{}") {
		t.Fatalf("expected route resolver constructor to return unified resolver:\n%s", text)
	}
	if !strings.Contains(text, "framework.PageOnlyRouteHandler") {
		t.Fatalf("expected page-only route handlers:\n%s", text)
	}
	if strings.Contains(text, "PageAndLiveRouteHandler") {
		t.Fatalf("did not expect live route handlers:\n%s", text)
	}
	if strings.Contains(text, "/.live/") {
		t.Fatalf("did not expect live route patterns:\n%s", text)
	}
	if strings.Contains(text, "ParseRootLiveState") {
		t.Fatalf("did not expect live resolver contract references:\n%s", text)
	}
	if !strings.Contains(text, "func NotFoundPage(notFound framework.NotFoundContext) templ.Component") {
		t.Fatalf("expected generated NotFoundPage helper in registry:\n%s", text)
	}
	if !strings.Contains(text, "RootLayout: r_root_root.RootLayout") {
		t.Fatalf("expected RootLayout wiring in page module:\n%s", text)
	}
	if !strings.Contains(text, "MetaGenChain: []framework.PageMetaGen") {
		t.Fatalf("expected generated metadata chain in page module:\n%s", text)
	}
	if !strings.Contains(text, "ErrorPage: func(locale string, path string) templ.Component") {
		t.Fatalf("expected generated ErrorPage fallback in page module:\n%s", text)
	}
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
		bundler.ProjectLayout{GeneratedImport: "internal/web/gen", AppModulePath: testAppModulePath},
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
	if err == nil {
		t.Fatal("expected missing root 404 metadata error")
	}
	if !strings.Contains(err.Error(), "missing root 404") {
		t.Fatalf("unexpected error: %v", err)
	}
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
		bundler.ProjectLayout{GeneratedImport: "internal/web/gen", AppModulePath: testAppModulePath},
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
	if err == nil {
		t.Fatal("expected missing root error metadata error")
	}
	if !strings.Contains(err.Error(), "missing root error") {
		t.Fatalf("unexpected error: %v", err)
	}
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
		bundler.ProjectLayout{GeneratedImport: "internal/web/gen", AppModulePath: testAppModulePath},
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
	if err != nil {
		t.Fatalf("generate registry: %v", err)
	}

	text := string(registry)
	if !strings.Contains(text, "component := r_error_author_param_slug.Error(view, pathValue)") {
		t.Fatalf("expected nearest author error template wiring, got:\n%s", text)
	}
}

func TestRewritePackageDeclarationAddsGeneratedMarker(t *testing.T) {
	source := "package appsrc\n\nimport (\n\t\"fmt\"\n)\n"

	rewritten, err := rewritePackageDeclaration([]byte(source), "r_page_root")
	if err != nil {
		t.Fatalf("rewrite package declaration: %v", err)
	}

	text := string(rewritten)
	if !strings.HasPrefix(text, "package r_page_root\n"+generatedTemplHeader+"\n") {
		t.Fatalf("expected generated marker after package declaration, got:\n%s", text)
	}
	if strings.Count(text, generatedTemplHeader) != 1 {
		t.Fatalf("expected exactly one generated marker, got:\n%s", text)
	}
}

func TestRewritePackageDeclarationKeepsSingleGeneratedMarker(t *testing.T) {
	source := "package appsrc\n\n" + generatedTemplHeader + "\n\ntempl Page() { <div></div> }\n"

	rewritten, err := rewritePackageDeclaration([]byte(source), "r_page_root")
	if err != nil {
		t.Fatalf("rewrite package declaration: %v", err)
	}

	text := string(rewritten)
	if strings.Count(text, generatedTemplHeader) != 1 {
		t.Fatalf("expected exactly one generated marker, got:\n%s", text)
	}
	if !strings.HasPrefix(text, "package r_page_root\n") {
		t.Fatalf("expected package rename to be applied, got:\n%s", text)
	}
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

import "example.com/app/internal/web/runtime"

templ Page(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "error.templ"), `package appsrc

import "example.com/app/internal/web/runtime"

templ Error(view runtime.RootLayoutView, path string) { <div>{ path }</div> }
`)
	writeTestFile(t, filepath.Join(appDir, "page.templ"), `package appsrc

import "example.com/app/internal/web/runtime"

templ Page(view runtime.NotesPageView) { <div>notes</div> }
`)

	err := Run(Config{
		Layout: bundler.ProjectLayout{
			RootDir:                 rootDir,
			AppDir:                  appDir,
			GeneratedDir:            genDir,
			GeneratedImport:         "internal/web/gen",
			ResolverDir:             resolverDir,
			AppModulePath:           testAppModulePath,
			PublicDir:               filepath.Join(rootDir, "public"),
			PublicRequestPathPrefix: "/",
		},
	})
	if err != nil {
		t.Fatalf("run approutegen: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(genDir, "registry_gen.go")); statErr != nil {
		t.Fatalf("expected generated registry file: %v", statErr)
	}
	serverSourcePath := filepath.Join(genDir, generatedServerFileName)
	if _, statErr := os.Stat(serverSourcePath); statErr != nil {
		t.Fatalf("expected generated server bootstrap: %v", statErr)
	}
	serverSource, err := os.ReadFile(serverSourcePath)
	if err != nil {
		t.Fatalf("read generated server bootstrap: %v", err)
	}
	serverText := string(serverSource)
	if strings.Contains(serverText, "appcore") {
		t.Fatalf("generated server bootstrap should not reference appcore:\n%s", serverText)
	}
	if !strings.Contains(serverText, `"example.com/app/internal/web/runtime"`) {
		t.Fatalf("generated server bootstrap should import runtime package:\n%s", serverText)
	}
	if !strings.Contains(serverText, "func NewHandler(cfg ServerConfig) (http.Handler, error)") {
		t.Fatalf("generated server bootstrap should expose NewHandler:\n%s", serverText)
	}
	if !strings.Contains(serverText, "func MountRoutes(mux *http.ServeMux, cfg ServerConfig) error") {
		t.Fatalf("generated server bootstrap should expose MountRoutes:\n%s", serverText)
	}
	if _, statErr := os.Stat(filepath.Join(resolverDir, generatedResolverFileName)); statErr != nil {
		t.Fatalf("expected generated resolver namespace: %v", statErr)
	}
}

func TestGenerateServerSourceHonorsFeatureFlags(t *testing.T) {
	disabledSource, err := generateServerSource(bundler.ProjectLayout{
		AppModulePath:           testAppModulePath,
		PublicRequestPathPrefix: "/public",
		ServerFeatures: bundler.ServerFeatures{
			I18nRouting:  false,
			StaticAssets: false,
			PublicFiles:  false,
		},
	})
	if err != nil {
		t.Fatalf("generate disabled server source: %v", err)
	}

	disabledText := string(disabledSource)
	if strings.Contains(disabledText, "frameworki18n") {
		t.Fatalf("disabled server source should omit i18n middleware wiring:\n%s", disabledText)
	}
	if strings.Contains(disabledText, "type StaticAssetsConfig struct") {
		t.Fatalf("disabled server source should omit static manifest config:\n%s", disabledText)
	}
	if strings.Contains(disabledText, "loadStaticMount(") {
		t.Fatalf("disabled server source should omit static loader helper:\n%s", disabledText)
	}
	if strings.Contains(disabledText, "type PublicFilesConfig struct") {
		t.Fatalf("disabled server source should omit public files config:\n%s", disabledText)
	}
	if strings.Contains(disabledText, "WithPublicFiles") {
		t.Fatalf("disabled server source should omit public files middleware:\n%s", disabledText)
	}

	enabledSource, err := generateServerSource(bundler.ProjectLayout{
		AppModulePath:           testAppModulePath,
		PublicRequestPathPrefix: "/public",
		ServerFeatures: bundler.ServerFeatures{
			I18nRouting:    true,
			StaticAssets:   true,
			PublicFiles:    true,
			HealthEndpoint: true,
		},
	})
	if err != nil {
		t.Fatalf("generate enabled server source: %v", err)
	}

	enabledText := string(enabledSource)
	if strings.Contains(enabledText, "appcore") {
		t.Fatalf("enabled server source should not reference appcore:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, `"example.com/app/internal/web/runtime"`) {
		t.Fatalf("enabled server source should import runtime package:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "frameworki18n.Middleware") {
		t.Fatalf("enabled server source should include i18n middleware:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "type RuntimeConfig struct") {
		t.Fatalf("enabled server source should include grouped runtime config:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "type Hooks struct") {
		t.Fatalf("enabled server source should include hook config:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "type StaticAssetsConfig struct") {
		t.Fatalf("enabled server source should include static manifest config:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "func loadStaticMount(manifestPath string, basePrefix string)") {
		t.Fatalf("enabled server source should include static loader helper:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "RequestPathPrefix string") {
		t.Fatalf("enabled server source should include public path prefix config:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "httpserver.WithPublicFiles") {
		t.Fatalf("enabled server source should include public files middleware:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "for idx := len(cfg.Hooks.Middleware) - 1; idx >= 0; idx--") {
		t.Fatalf(
			"enabled server source should wrap middleware hooks in reverse for request-order stability:\n%s",
			enabledText,
		)
	}
	if !strings.Contains(enabledText, "for _, mount := range cfg.Hooks.Mount") {
		t.Fatalf("enabled server source should include mount hooks:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "handler = frameworki18n.Middleware") {
		t.Fatalf("enabled server source should apply i18n after middleware hooks:\n%s", enabledText)
	}
	if !strings.Contains(enabledText, "versionedPrefix := manifest.VersionedURLPrefix(basePrefix)") {
		t.Fatalf("enabled server source should compose runtime static prefix from manifest hash:\n%s", enabledText)
	}
	if strings.Index(enabledText, "for idx := len(cfg.Hooks.Middleware) - 1; idx >= 0; idx--") >
		strings.Index(enabledText, "handler = frameworki18n.Middleware") {
		t.Fatalf("middleware hook wrapping should occur before i18n middleware:\n%s", enabledText)
	}
	if strings.Index(enabledText, "handler = frameworki18n.Middleware") >
		strings.Index(enabledText, "httpserver.WithPublicFiles") {
		t.Fatalf("public files middleware should be applied after i18n middleware:\n%s", enabledText)
	}
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(filePath), err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", filePath, err)
	}
}
