package templcssgen

import (
	"bytes"
	"fmt"
	gofmt "go/format"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/filesystem"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

type WriteStylesheetConfig struct {
	Layout     projectlayout.ProjectLayout
	OutputPath string
}

type PrepareStaticSourceConfig struct {
	Layout         projectlayout.ProjectLayout
	SourceDir      string
	StylesheetPath string
}

func WriteStylesheet(cfg WriteStylesheetConfig) error {
	layout, err := validateLayout(cfg.Layout)
	if err != nil {
		return err
	}

	outputPath := strings.TrimSpace(cfg.OutputPath)
	if outputPath == "" {
		return fmt.Errorf("stylesheet output path is required")
	}
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(layout.RootDir, outputPath)
	}
	outputPath = filepath.ToSlash(outputPath)

	tempDir, err := os.MkdirTemp(layout.RootDir, ".no-js-templ-css-*")
	if err != nil {
		return fmt.Errorf("create stylesheet helper dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	helperSource, err := stylesheetHelperSource(layout, outputPath)
	if err != nil {
		return err
	}

	helperPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(helperPath, helperSource, 0o644); err != nil {
		return fmt.Errorf("write stylesheet helper %q: %w", filepath.ToSlash(helperPath), err)
	}

	relativeDir, err := filepath.Rel(layout.RootDir, tempDir)
	if err != nil {
		return fmt.Errorf("resolve helper dir relative path: %w", err)
	}
	goRunPath := "./" + filepath.ToSlash(relativeDir)
	goRunPath = strings.TrimSuffix(goRunPath, "/")

	cmd := exec.Command("go", "run", goRunPath)
	cmd.Dir = layout.RootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			return fmt.Errorf("run stylesheet helper: %w", err)
		}
		return fmt.Errorf("run stylesheet helper: %w: %s", err, trimmed)
	}

	return nil
}

func PrepareStaticSource(cfg PrepareStaticSourceConfig) (string, func() error, error) {
	layout, err := validateLayout(cfg.Layout)
	if err != nil {
		return "", nil, err
	}

	sourceDir := strings.TrimSpace(cfg.SourceDir)
	if sourceDir == "" {
		return "", nil, fmt.Errorf("static asset source dir is required")
	}
	if !filepath.IsAbs(sourceDir) {
		sourceDir = filepath.Join(layout.RootDir, sourceDir)
	}
	sourceDir = filepath.ToSlash(sourceDir)

	if err := Run(Config{Layout: layout}); err != nil {
		return "", nil, err
	}

	stageDir, err := os.MkdirTemp("", "no-js-static-assets-stage-*")
	if err != nil {
		return "", nil, fmt.Errorf("create static asset stage dir: %w", err)
	}
	cleanup := func() error {
		return os.RemoveAll(stageDir)
	}

	if err := copyStaticSourceIfPresent(sourceDir, stageDir); err != nil {
		_ = cleanup()
		return "", nil, fmt.Errorf("copy static assets to stage dir: %w", err)
	}

	stylesheetPath := strings.TrimSpace(cfg.StylesheetPath)
	if stylesheetPath == "" {
		stylesheetPath = defaultStylesheetAssetPath
	}
	if err := WriteStylesheet(WriteStylesheetConfig{
		Layout:     layout,
		OutputPath: filepath.Join(stageDir, filepath.FromSlash(stylesheetPath)),
	}); err != nil {
		_ = cleanup()
		return "", nil, err
	}

	return stageDir, cleanup, nil
}

func copyStaticSourceIfPresent(sourceDir string, stageDir string) error {
	info, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(stageDir, 0o755)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", sourceDir)
	}

	return filesystem.CopyTree(sourceDir, stageDir)
}

func stylesheetHelperSource(layout projectlayout.ProjectLayout, outputPath string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteString("package main\n\n")
	buffer.WriteString("import (\n")
	buffer.WriteString("\t\"fmt\"\n")
	buffer.WriteString("\t\"os\"\n")
	buffer.WriteString("\t\"path/filepath\"\n")
	buffer.WriteString("\t\"strings\"\n")
	buffer.WriteString("\t\"github.com/a-h/templ\"\n")
	buffer.WriteString("\tgen " + quote(path.Join(layout.AppModulePath, layout.GeneratedImport)) + "\n")
	buffer.WriteString(")\n\n")
	buffer.WriteString("func main() {\n")
	buffer.WriteString("\tclasses := templ.NewCSSHandler(gen.TemplCSSClasses()...).Classes\n")
	buffer.WriteString("\tseen := make(map[string]struct{}, len(classes))\n")
	buffer.WriteString("\tvar builder strings.Builder\n")
	buffer.WriteString("\tfor _, class := range classes {\n")
	buffer.WriteString("\t\tif _, ok := seen[class.ID]; ok {\n")
	buffer.WriteString("\t\t\tcontinue\n")
	buffer.WriteString("\t\t}\n")
	buffer.WriteString("\t\tseen[class.ID] = struct{}{}\n")
	buffer.WriteString("\t\tbuilder.WriteString(string(class.Class))\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\toutputPath := " + quote(outputPath) + "\n")
	buffer.WriteString("\tif err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {\n")
	buffer.WriteString("\t\tpanic(fmt.Errorf(\"create stylesheet dir: %w\", err))\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tif err := os.WriteFile(outputPath, []byte(builder.String()), 0o644); err != nil {\n")
	buffer.WriteString("\t\tpanic(fmt.Errorf(\"write stylesheet: %w\", err))\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n")

	formatted, err := gofmt.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format stylesheet helper source: %w", err)
	}
	return formatted, nil
}
