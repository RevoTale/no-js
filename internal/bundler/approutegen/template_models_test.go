package approutegen

import (
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

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
