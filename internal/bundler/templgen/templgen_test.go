package templgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunGeneratesOutputFromPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourcePath := filepath.Join(root, "hello.templ")
	writeTestFile(t, sourcePath, "package sample\n\ntempl Hello() { <div>hello</div> }\n")

	err := Run(Config{
		Paths:    []string{root},
		BasePath: root,
	})
	require.NoError(t, err)

	generatedPath := filepath.Join(root, "hello_templ.go")
	_, statErr := os.Stat(generatedPath)
	require.NoError(t, statErr)
}

func TestRunReturnsErrorForEmptySelection(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := Run(Config{
		BasePath: root,
	})
	require.Error(t, err)
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}
