package approutegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/clientassetext"
)

func writeTemplCopy(tpl templateDef) error {
	source, err := os.ReadFile(tpl.SourcePath)
	if err != nil {
		return fmt.Errorf("read %q: %w", tpl.SourcePath, err)
	}

	rewritten, err := rewritePackageDeclaration(source, tpl.Package)
	if err != nil {
		return fmt.Errorf("rewrite package for %q: %w", tpl.SourcePath, err)
	}

	if err := os.MkdirAll(tpl.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir %q: %w", tpl.OutputDir, err)
	}

	target := filepath.Join(tpl.OutputDir, tpl.OutputFile)
	if err := os.WriteFile(target, rewritten, 0o644); err != nil {
		return fmt.Errorf("write generated template %q: %w", target, err)
	}

	return nil
}

func rewritePackageDeclaration(source []byte, packageName string) ([]byte, error) {
	lines := strings.Split(string(source), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "package ") {
			continue
		}

		lines[i] = "package " + packageName

		firstNonEmpty := i + 1
		for firstNonEmpty < len(lines) && strings.TrimSpace(lines[firstNonEmpty]) == "" {
			firstNonEmpty++
		}
		if firstNonEmpty < len(lines) && strings.TrimSpace(lines[firstNonEmpty]) == generatedTemplHeader {
			return []byte(strings.Join(lines, "\n")), nil
		}

		rewritten := make([]string, 0, len(lines)+1)
		rewritten = append(rewritten, lines[:i+1]...)
		rewritten = append(rewritten, generatedTemplHeader)
		rewritten = append(rewritten, lines[i+1:]...)
		return []byte(strings.Join(rewritten, "\n")), nil
	}

	return nil, errors.New("template missing package declaration")
}

func rewriteGoPackageDeclaration(source []byte, packageName string) ([]byte, error) {
	lines := strings.Split(string(source), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "package ") {
			continue
		}
		lines[idx] = "package " + packageName
		rewritten := strings.Join(lines, "\n")
		if strings.Contains(rewritten, generatedGoHeader) {
			return []byte(rewritten), nil
		}
		return []byte(generatedGoHeader + "\n" + rewritten), nil
	}
	return nil, errors.New("go source missing package declaration")
}

func writeClientAssetHelperCopies(tpl templateDef) error {
	if strings.TrimSpace(tpl.SourcePath) == "" || strings.TrimSpace(tpl.OutputDir) == "" {
		return nil
	}
	entries, err := os.ReadDir(filepath.Dir(tpl.SourcePath))
	if err != nil {
		return fmt.Errorf("read client asset helper dir for %q: %w", tpl.SourcePath, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !isClientAssetHelperFile(entry.Name()) {
			continue
		}
		sourcePath := filepath.Join(filepath.Dir(tpl.SourcePath), entry.Name())
		source, readErr := os.ReadFile(sourcePath)
		if readErr != nil {
			return fmt.Errorf("read %q: %w", sourcePath, readErr)
		}
		rewritten, rewriteErr := rewriteGoPackageDeclaration(source, tpl.Package)
		if rewriteErr != nil {
			return fmt.Errorf("rewrite package for %q: %w", sourcePath, rewriteErr)
		}
		outPath := filepath.Join(tpl.OutputDir, entry.Name())
		if writeErr := os.WriteFile(outPath, rewritten, 0o644); writeErr != nil {
			return fmt.Errorf("write %q: %w", outPath, writeErr)
		}
	}
	return nil
}

func isClientAssetHelperFile(name string) bool {
	return clientassetext.IsGeneratedHelperName(name)
}

func writeSourcePackageCopy(sourcePackage sourcePackageDef, generatedRoot string) error {
	outputDir := filepath.Join(generatedRoot, sourcePackage.ModuleName)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create source package output %q: %w", outputDir, err)
	}

	for _, filePath := range sourcePackage.Files {
		source, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %q: %w", filePath, err)
		}
		rewritten, err := rewriteGoPackageDeclaration(source, sourcePackage.Package)
		if err != nil {
			return fmt.Errorf("rewrite package for %q: %w", filePath, err)
		}
		outPath := filepath.Join(outputDir, filepath.Base(filePath))
		if err := os.WriteFile(outPath, rewritten, 0o644); err != nil {
			return fmt.Errorf("write %q: %w", outPath, err)
		}
	}

	paramsSource, err := generateSourcePackageParamsFile(sourcePackage)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "params_gen.go"), paramsSource, 0o644); err != nil {
		return fmt.Errorf("write source params file for %q: %w", sourcePackage.InternalRouteID, err)
	}

	return nil
}

func generateSourcePackageParamsFile(sourcePackage sourcePackageDef) ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteString(generatedGoHeader + "\n")
	writef(buffer, "package %s\n\n", sourcePackage.Package)
	imports := []string{fmt.Sprintf("%q", frameworkModulePath+"/framework/router")}
	if len(sourcePackage.Params) > 0 {
		imports = append(imports, "\"strings\"")
	}
	for _, param := range sourcePackage.Params {
		if param.Type == "[]string" {
			imports = append(imports, "\"slices\"")
			break
		}
	}
	buffer.WriteString("import (\n")
	for _, line := range dedupeSorted(imports) {
		buffer.WriteString("\t" + line + "\n")
	}
	buffer.WriteString(")\n\n")
	writef(buffer, "type %s struct {\n", sourcePackage.ParamsTypeName)
	if len(sourcePackage.Params) == 0 {
		buffer.WriteString("}\n")
	} else {
		for _, param := range sourcePackage.Params {
			writef(buffer, "\t%s %s\n", param.FieldName, param.Type)
		}
		buffer.WriteString("}\n")
	}
	buffer.WriteString("\n")

	writef(buffer, "func ParseParams(requestPath string) (%s, bool) {\n", sourcePackage.ParamsTypeName)
	if len(sourcePackage.Params) == 0 {
		writef(buffer, "\t_, ok := router.MatchPathPattern(%q, requestPath)\n", routePattern(sourcePackage.PublicRouteID))
	} else {
		writef(
			buffer,
			"\tparams, ok := router.MatchPathPattern(%q, requestPath)\n",
			routePattern(sourcePackage.PublicRouteID),
		)
	}
	buffer.WriteString("\tif !ok {\n")
	writef(buffer, "\t\treturn %s{}, false\n", sourcePackage.ParamsTypeName)
	buffer.WriteString("\t}\n")
	writef(buffer, "\tout := %s{}\n", sourcePackage.ParamsTypeName)
	for _, param := range sourcePackage.Params {
		switch param.Type {
		case "[]string":
			writef(buffer, "\tif %sValue, exists := params[%q]; exists {\n", param.FieldName, param.Name)
			writef(buffer, "\t\tout.%s = slices.Clone(%sValue)\n", param.FieldName, param.FieldName)
			buffer.WriteString("\t}\n")
		default:
			writef(buffer, "\t%sValue, exists := params[%q]\n", param.FieldName, param.Name)
			buffer.WriteString("\tif !exists || len(" + param.FieldName + "Value) == 0 {\n")
			writef(buffer, "\t\treturn %s{}, false\n", sourcePackage.ParamsTypeName)
			buffer.WriteString("\t}\n")
			writef(buffer, "\tout.%s = strings.TrimSpace(%sValue[0])\n", param.FieldName, param.FieldName)
		}
	}
	buffer.WriteString("\treturn out, true\n")
	buffer.WriteString("}\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format source params for %q: %w", sourcePackage.InternalRouteID, err)
	}
	return formatted, nil
}
