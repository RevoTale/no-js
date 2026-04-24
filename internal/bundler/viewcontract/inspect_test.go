package viewcontract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInspectDetectsZeroArgSystemPagesAndOptionalHooks(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

import "github.com/a-h/templ"

type RootLayoutView struct{}

func NewNotFoundView() RootLayoutView { return RootLayoutView{} }
func NewErrorView() RootLayoutView { return RootLayoutView{} }
func SetStaticAssetBasePath(prefix string) {}
func TemplCSSVariants() []templ.CSSClass { return nil }
`)

	inspection, err := Inspect(viewDir)
	require.NoError(t, err)

	require.True(t, inspection.SystemPages.HasZeroArgNotFoundView)
	require.True(t, inspection.SystemPages.HasZeroArgErrorView)
	require.True(t, inspection.HasStaticAssetBasePathHook)
	require.True(t, inspection.HasTemplCSSVariantsHook)

	require.NoError(t, inspection.SystemPages.Validate())
}

func TestSystemPageValidationRejectsTypedConstructors(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type RootLayoutView struct{}

func NewNotFoundView(messages *Messages) RootLayoutView { return RootLayoutView{} }
func NewErrorView(messages *Messages) RootLayoutView { return RootLayoutView{} }
`)

	inspection, err := Inspect(viewDir)
	require.NoError(t, err)

	err = inspection.SystemPages.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "web/view must define func NewNotFoundView() RootLayoutView")
}

func TestInspectRejectsNonViewPackageName(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "context.go"), `package runtime

type Context struct{}
`)

	_, err := Inspect(viewDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "web/view files must declare package view")
	require.Contains(t, err.Error(), "declares package runtime")
}

func TestSystemPageValidationAllowsMissingErrorConstructor(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type RootLayoutView struct{}

func NewNotFoundView() RootLayoutView { return RootLayoutView{} }
`)

	inspection, err := Inspect(viewDir)
	require.NoError(t, err)

	require.NoError(t, inspection.SystemPages.Validate())
}

func TestSystemPageValidationRejectsMissingNotFoundConstructor(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type RootLayoutView struct{}

func NewErrorView() RootLayoutView { return RootLayoutView{} }
`)

	inspection, err := Inspect(viewDir)
	require.NoError(t, err)

	err = inspection.SystemPages.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "web/view must define func NewNotFoundView() RootLayoutView")
}

func writeViewFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
