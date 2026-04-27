package approutegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RevoTale/no-js/internal/bundler/viewcontract"
	"github.com/stretchr/testify/require"
)

const testAppModulePath = "example.com/app"

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
