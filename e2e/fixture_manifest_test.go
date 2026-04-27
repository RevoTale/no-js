package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixtureAppsDeclareToolsAndLocalReplaceInGoMod(t *testing.T) {
	t.Helper()

	fixturesDir := filepath.Join(repoRootPath(t), "e2e", "testdata")
	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		goModPath := filepath.Join(fixturesDir, entry.Name(), "go.mod")
		goMod, err := os.ReadFile(goModPath)
		require.NoError(t, err, "%s missing go.mod", entry.Name())

		contents := string(goMod)
		require.Contains(t, contents, "module example.com/no-js-e2e/"+entry.Name())
		require.Contains(t, contents, "tool (")
		require.Contains(t, contents, "github.com/RevoTale/no-js/cmd/no-js")
		require.Contains(t, contents, "github.com/RevoTale/no-js/cmd/templgen")
		require.Contains(t, contents, "github.com/evanw/esbuild v0.28.0 // indirect")
		require.Contains(t, contents, "github.com/nicksnyder/go-i18n/v2 v2.6.1 // indirect")
		require.Contains(t, contents, "github.com/tdewolff/parse/v2 v2.8.12 // indirect")
		require.Contains(t, contents, "replace github.com/RevoTale/no-js => ../../..")

		goSumPath := filepath.Join(fixturesDir, entry.Name(), "go.sum")
		goSum, err := os.ReadFile(goSumPath)
		require.NoError(t, err, "%s missing go.sum", entry.Name())
		require.Contains(t, string(goSum), "github.com/evanw/esbuild v0.28.0/go.mod")
		require.Contains(t, string(goSum), "github.com/tdewolff/parse/v2 v2.8.12/go.mod")

		serverPath := filepath.Join(fixturesDir, entry.Name(), "cmd", "server", "main.go")
		_, err = os.Stat(serverPath)
		require.NoError(t, err, "%s missing cmd/server/main.go", entry.Name())

		probePath := filepath.Join(fixturesDir, entry.Name(), "cmd", "probe", "main.go")
		_, err = os.Stat(probePath)
		require.ErrorIs(t, err, os.ErrNotExist, "%s still has cmd/probe/main.go", entry.Name())
	}
}
