package viewcontract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInspectDetectsOptionalHooks(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

import "github.com/a-h/templ"

func SetStaticAssetBasePath(prefix string) {}
func TemplCSSVariants() []templ.CSSClass { return nil }
`)

	inspection, err := Inspect(viewDir)
	require.NoError(t, err)

	require.True(t, inspection.HasStaticAssetBasePathHook)
	require.True(t, inspection.HasTemplCSSVariantsHook)
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

func TestInspectDoesNotRequireSystemPageConstructors(t *testing.T) {
	t.Parallel()

	viewDir := t.TempDir()
	writeViewFile(t, filepath.Join(viewDir, "view_models.go"), `package view

type HomePageView struct{}
`)

	_, err := Inspect(viewDir)
	require.NoError(t, err)
}

func writeViewFile(t *testing.T, filePath string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
