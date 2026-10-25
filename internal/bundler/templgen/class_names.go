package templgen

import (
	"bytes"
	"go/ast"
	"go/format"
	goparser "go/parser"
	"go/token"
	"strconv"
)

// optimizeClassNames changes only templ's generated class-list string conversion.
// CSS registration, escaping, and evaluation of each class expression stay intact.
func optimizeClassNames(source []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, "generated.go", source, goparser.ParseComments)
	if err != nil {
		return nil, err
	}
	identifiers := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if name, ok := node.(*ast.Ident); ok {
			identifiers[name.Name] = true
		}
		return true
	})
	alias := "templ_nojs_runtime"
	for identifiers[alias] {
		alias += "_"
	}
	changed := false
	ast.Inspect(file, func(node ast.Node) bool {
		if declaration, ok := node.(*ast.FuncDecl); ok {
			return generatedTemplateFunction(declaration)
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || method.Sel.Name != "String" || len(call.Args) != 0 {
			return true
		}
		classes, ok := method.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		factory, ok := classes.Fun.(*ast.SelectorExpr)
		if !ok || (factory.Sel.Name != "Classes" && factory.Sel.Name != "CSSClasses") {
			return true
		}
		ellipsis := classes.Ellipsis
		if factory.Sel.Name == "CSSClasses" {
			if len(classes.Args) != 1 {
				return true
			}
			ellipsis = classes.Rparen
		}
		pkg, ok := factory.X.(*ast.Ident)
		// A resolved object is a user declaration shadowing the templ import.
		if !ok || pkg.Name != "templ" || pkg.Obj != nil {
			return true
		}
		*call = ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: ast.NewIdent(alias), Sel: ast.NewIdent("ClassNames")},
			Args: classes.Args, Ellipsis: ellipsis,
		}
		changed = true
		return true
	})
	if !changed {
		return source, nil
	}
	spec := &ast.ImportSpec{
		Name: ast.NewIdent(alias),
		Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("github.com/RevoTale/no-js/v2/framework/templruntime")},
	}
	file.Decls = append([]ast.Decl{&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{spec}}}, file.Decls...)
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func generatedTemplateFunction(declaration *ast.FuncDecl) bool {
	if declaration.Body == nil {
		return false
	}
	for _, statement := range declaration.Body.List {
		returned, ok := statement.(*ast.ReturnStmt)
		if !ok || len(returned.Results) != 1 {
			continue
		}
		call, ok := returned.Results[0].(*ast.CallExpr)
		if !ok {
			continue
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "GeneratedTemplate" {
			continue
		}
		pkg, ok := selector.X.(*ast.Ident)
		if ok && pkg.Name == "templruntime" {
			return true
		}
	}
	return false
}
