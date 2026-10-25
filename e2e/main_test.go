package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if noJSToolPath != "" {
		if err := os.RemoveAll(filepath.Dir(noJSToolPath)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "clean up e2e CLI: %v\n", err)
			code = 1
		}
	}
	os.Exit(code)
}
