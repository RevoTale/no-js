package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

func CopyTree(sourceRoot string, targetRoot string) error {
	return filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return fmt.Errorf("resolve relative path for %q: %w", path, err)
		}
		if relativePath == "." {
			return os.MkdirAll(targetRoot, 0o755)
		}

		targetPath := filepath.Join(targetRoot, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return err
		}

		return nil
	})
}
