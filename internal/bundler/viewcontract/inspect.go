package viewcontract

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Inspection struct {
	HasStaticAssetBasePathHook bool
	HasTemplCSSVariantsHook    bool
}

func Inspect(viewDir string) (Inspection, error) {
	trimmedDir := strings.TrimSpace(viewDir)
	if trimmedDir == "" {
		return Inspection{}, fmt.Errorf("view dir is required")
	}

	entries, err := os.ReadDir(trimmedDir)
	if err != nil {
		return Inspection{}, fmt.Errorf("read view dir %q: %w", trimmedDir, err)
	}

	inspection := Inspection{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(trimmedDir, name)
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return Inspection{}, fmt.Errorf("parse view file %q: %w", filepath.ToSlash(filePath), err)
		}
		packageName := ""
		if file.Name != nil {
			packageName = file.Name.Name
		}
		if packageName != "view" {
			return Inspection{}, fmt.Errorf(
				"web/view files must declare package view: %s declares package %s",
				filepath.ToSlash(filePath),
				packageName,
			)
		}
		inspectFile(&inspection, fset, file)
	}

	return inspection, nil
}

func inspectFile(inspection *Inspection, fset *token.FileSet, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}

		switch fn.Name.Name {
		case "SetStaticAssetBasePath":
			inspection.HasStaticAssetBasePathHook = matchesStaticAssetBasePathHook(fset, fn)
		case "TemplCSSVariants":
			inspection.HasTemplCSSVariantsHook = matchesTemplCSSVariantsHook(fset, fn)
		}
	}
}

func matchesStaticAssetBasePathHook(fset *token.FileSet, fn *ast.FuncDecl) bool {
	if fn.Recv != nil {
		return false
	}
	if countFields(fn.Type.Params) != 1 {
		return false
	}
	paramType, ok := singleParamType(fset, fn.Type.Params)
	if !ok || paramType != "string" {
		return false
	}
	return countFields(fn.Type.Results) == 0
}

func matchesTemplCSSVariantsHook(fset *token.FileSet, fn *ast.FuncDecl) bool {
	if fn.Recv != nil {
		return false
	}
	if countFields(fn.Type.Params) != 0 {
		return false
	}
	return singleResultType(fset, fn.Type.Results) == "[]templ.CSSClass"
}

func singleParamType(fset *token.FileSet, fields *ast.FieldList) (string, bool) {
	if countFields(fields) != 1 || len(fields.List) != 1 {
		return "", false
	}
	return normalizeExprString(fset, fields.List[0].Type), true
}

func singleResultType(fset *token.FileSet, fields *ast.FieldList) string {
	if countFields(fields) != 1 || len(fields.List) != 1 {
		return ""
	}
	return normalizeExprString(fset, fields.List[0].Type)
}

func countFields(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}

	count := 0
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += len(field.Names)
	}
	return count
}

func normalizeExprString(fset *token.FileSet, expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	var buffer bytes.Buffer
	if err := format.Node(&buffer, fset, expr); err != nil {
		return ""
	}
	return strings.Join(strings.Fields(buffer.String()), "")
}
