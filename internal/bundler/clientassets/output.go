package clientassets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/esbuildtarget"
	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/evanw/esbuild/pkg/api"
)

func writeStaticOutputs(stageDir string, layout projectlayout.ProjectLayout, plan Plan) error {
	for _, bundle := range plan.cssBundles {
		if err := writeCSSBundle(stageDir, bundle); err != nil {
			return err
		}
	}
	if err := writeScriptBundles(stageDir, layout, plan.scriptFiles); err != nil {
		return err
	}
	return nil
}

func writeCSSBundle(stageDir string, bundle *cssBundle) error {
	if bundle == nil || len(bundle.CSSFiles) == 0 {
		return nil
	}
	builder := strings.Builder{}
	for _, file := range bundle.CSSFiles {
		if file == nil || strings.TrimSpace(file.Transformed) == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(file.Transformed)
		if !strings.HasSuffix(file.Transformed, "\n") {
			builder.WriteByte('\n')
		}
	}
	if strings.TrimSpace(builder.String()) == "" {
		return nil
	}
	return writeStageFile(stageDir, bundle.AssetPath, []byte(builder.String()))
}

func writeScriptBundles(stageDir string, layout projectlayout.ProjectLayout, files []*scriptAsset) error {
	if len(files) == 0 {
		return nil
	}
	entryPoints := make([]string, 0, len(files))
	assetPathOwners := map[string]string{}
	for _, file := range files {
		if file == nil {
			continue
		}
		assetPath := filepath.ToSlash(file.AssetPath)
		if existing, ok := assetPathOwners[assetPath]; ok {
			return fmt.Errorf("client script output %q is produced by both %q and %q", assetPath, existing, file.SourcePath)
		}
		assetPathOwners[assetPath] = file.SourcePath
		entryPoints = append(entryPoints, file.SourcePath)
	}
	if len(entryPoints) == 0 {
		return nil
	}

	target, engines, err := esbuildtarget.Parse(layout.Assets.BrowserTargets)
	if err != nil {
		return fmt.Errorf("parse browser targets: %w", err)
	}

	webRoot := filepath.Dir(layout.RoutesDir)
	result := api.Build(api.BuildOptions{
		EntryPoints:       entryPoints,
		AbsWorkingDir:     layout.RootDir,
		Outdir:            stageDir,
		Outbase:           webRoot,
		EntryNames:        "[dir]/[name]",
		ChunkNames:        "chunks/[name]-[hash]",
		Bundle:            true,
		Write:             true,
		Format:            api.FormatESModule,
		Splitting:         true,
		Platform:          api.PlatformBrowser,
		Target:            target,
		Engines:           engines,
		MinifyWhitespace:  true,
		MinifySyntax:      true,
		MinifyIdentifiers: true,
	})
	if len(result.Errors) > 0 {
		return fmt.Errorf("bundle client scripts: %s", result.Errors[0].Text)
	}
	return nil
}

func writeStageFile(stageDir string, assetPath string, content []byte) error {
	target := filepath.Join(stageDir, filepath.FromSlash(strings.TrimLeft(assetPath, "/")))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create asset dir for %q: %w", target, err)
	}
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return fmt.Errorf("write asset %q: %w", target, err)
	}
	return nil
}

func writeCSSHelper(asset *cssAsset) error {
	builder := strings.Builder{}
	builder.WriteString(generatedHeader)
	builder.WriteString("\n\npackage ")
	builder.WriteString(asset.PackageName)
	builder.WriteString("\n\n")
	if len(asset.Classes) > 0 {
		builder.WriteString("const (\n")
		for _, class := range asset.Classes {
			builder.WriteString("\t")
			builder.WriteString(class.Constant)
			builder.WriteString(" = ")
			builder.WriteString(strconv.Quote(class.Generated))
			builder.WriteString("\n")
		}
		builder.WriteString(")\n")
	}
	return os.WriteFile(asset.HelperPath, []byte(builder.String()), 0o644)
}

func writeScriptHelper(asset *scriptAsset) error {
	builder := strings.Builder{}
	builder.WriteString(generatedHeader)
	builder.WriteString("\n\npackage ")
	builder.WriteString(asset.PackageName)
	builder.WriteString(`

import (
	"context"
	"html"
	"io"

	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/a-h/templ"
)

var `)
	builder.WriteString(asset.VarName)
	builder.WriteString(" = templ.NewOnceHandle(templ.WithComponent(templ.ComponentFunc(\n")
	builder.WriteString("\tfunc(ctx context.Context, w io.Writer) error {\n")
	builder.WriteString("\tsrc := metagen.AssetURL(ctx, ")
	builder.WriteString(strconv.Quote(asset.AssetPath))
	builder.WriteString(")\n")
	builder.WriteString("\tif src == \"\" {\n\t\treturn nil\n\t}\n")
	builder.WriteString("\t_, err := io.WriteString(w, `<script type=\"module\" src=\"`+\n")
	builder.WriteString("\t\thtml.EscapeString(src)+`\"></script>`)\n")
	builder.WriteString("\treturn err\n})))\n\n")
	builder.WriteString("func ")
	builder.WriteString(asset.FuncName)
	builder.WriteString("() templ.Component {\n\treturn ")
	builder.WriteString(asset.VarName)
	builder.WriteString(".Once()\n}\n")
	return os.WriteFile(asset.HelperPath, []byte(builder.String()), 0o644)
}

func cleanGeneratedHelpers(layout projectlayout.ProjectLayout) error {
	for _, root := range clientAssetRoots(layout) {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if !isGeneratedHelper(filePath) {
				return nil
			}
			content, readErr := os.ReadFile(filePath)
			if readErr != nil {
				return readErr
			}
			if strings.HasPrefix(string(content), generatedHeader) {
				return os.Remove(filePath)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
