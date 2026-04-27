package approutegen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/RevoTale/no-js/internal/projectlayout"
)

func validateDiscoveryConventions(paths projectlayout.ProjectLayout, discovery *discoveryConventions) error {
	if discovery == nil {
		return nil
	}

	if strings.TrimSpace(discovery.RobotsFile) != "" {
		if err := validateDiscoveryFunction(
			discovery.RobotsFile,
			"Robots",
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isRuntimeContextType(expr, imports, viewImportPath(paths))
				},
				func(expr ast.Expr, imports map[string]string) bool {
					return isPointerToImportedSelector(expr, imports, "net/http", "Request")
				},
			},
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isImportedSelector(expr, imports, frameworkModulePath+"/framework/discovery", "Robots")
				},
				isErrorType,
			},
		); err != nil {
			return err
		}
		discovery.HasRobots = true
	}

	for idx := range discovery.Sitemaps {
		convention := &discovery.Sitemaps[idx]
		if err := validateDiscoveryFunction(
			convention.SourcePath,
			"Sitemap",
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isRuntimeContextType(expr, imports, viewImportPath(paths))
				},
				func(expr ast.Expr, imports map[string]string) bool {
					return isPointerToImportedSelector(expr, imports, "net/http", "Request")
				},
			},
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isSliceOfImportedSelector(expr, imports, frameworkModulePath+"/framework/discovery", "SitemapEntry")
				},
				isErrorType,
			},
		); err != nil {
			return err
		}
		convention.HasSitemap = true

		generateErr := validateDiscoveryFunction(
			convention.SourcePath,
			"GenerateSitemaps",
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isRuntimeContextType(expr, imports, viewImportPath(paths))
				},
				func(expr ast.Expr, imports map[string]string) bool {
					return isPointerToImportedSelector(expr, imports, "net/http", "Request")
				},
			},
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isSliceOfImportedSelector(expr, imports, frameworkModulePath+"/framework/discovery", "SitemapID")
				},
				isErrorType,
			},
		)
		switch {
		case generateErr == nil:
			convention.HasGenerateSitemaps = true
		case !errors.Is(generateErr, fs.ErrNotExist):
			return generateErr
		}

		sitemapByIDErr := validateDiscoveryFunction(
			convention.SourcePath,
			"SitemapChunk",
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isRuntimeContextType(expr, imports, viewImportPath(paths))
				},
				func(expr ast.Expr, imports map[string]string) bool {
					return isPointerToImportedSelector(expr, imports, "net/http", "Request")
				},
				isStringType,
			},
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isSliceOfImportedSelector(expr, imports, frameworkModulePath+"/framework/discovery", "SitemapEntry")
				},
				isErrorType,
			},
		)
		switch {
		case sitemapByIDErr == nil:
			convention.HasSitemapChunk = true
		case !errors.Is(sitemapByIDErr, fs.ErrNotExist):
			return sitemapByIDErr
		}

		hasDynamicSitemapPart := convention.HasGenerateSitemaps || convention.HasSitemapChunk
		if hasDynamicSitemapPart && (!convention.HasGenerateSitemaps || !convention.HasSitemapChunk) {
			return fmt.Errorf(
				"%s: dynamic sitemaps require GenerateSitemaps and SitemapChunk",
				convention.SourcePath,
			)
		}
	}
	sortSitemapConventions(discovery.Sitemaps)
	if err := validateDiscoveryPatternConflicts(discovery.Sitemaps, func(convention sitemapConvention) string {
		return conventionEndpointPattern(convention.RouteID, "sitemap.xml")
	}); err != nil {
		return err
	}

	for _, convention := range discovery.Feeds {
		if err := validateDiscoveryFunction(
			convention.SourcePath,
			"Feed",
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isRuntimeContextType(expr, imports, viewImportPath(paths))
				},
				func(expr ast.Expr, imports map[string]string) bool {
					return isPointerToImportedSelector(expr, imports, "net/http", "Request")
				},
			},
			[]func(ast.Expr, map[string]string) bool{
				func(expr ast.Expr, imports map[string]string) bool {
					return isImportedSelector(expr, imports, frameworkModulePath+"/framework/discovery", "FeedDocument")
				},
				isErrorType,
			},
		); err != nil {
			return err
		}
	}
	sortFeedConventions(discovery.Feeds)
	if err := validateDiscoveryPatternConflicts(discovery.Feeds, func(convention feedConvention) string {
		return conventionEndpointPattern(convention.RouteID, "feed.xml")
	}); err != nil {
		return err
	}

	return nil
}

func validateDiscoveryFunction(
	filePath string,
	functionName string,
	paramChecks []func(ast.Expr, map[string]string) bool,
	resultChecks []func(ast.Expr, map[string]string) bool,
) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}

	imports := importAliases(file)
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != functionName {
			continue
		}
		if funcDecl.Recv != nil {
			return fmt.Errorf("%s: %s must be a package function", filePath, functionName)
		}
		paramTypes := fieldTypes(funcDecl.Type.Params)
		if len(paramTypes) != len(paramChecks) {
			return fmt.Errorf("%s: %s has invalid parameter count", filePath, functionName)
		}
		for idx, check := range paramChecks {
			if !check(paramTypes[idx], imports) {
				return fmt.Errorf("%s: %s has invalid parameter %d signature", filePath, functionName, idx+1)
			}
		}

		resultTypes := fieldTypes(funcDecl.Type.Results)
		if len(resultTypes) != len(resultChecks) {
			return fmt.Errorf("%s: %s has invalid result count", filePath, functionName)
		}
		for idx, check := range resultChecks {
			if !check(resultTypes[idx], imports) {
				return fmt.Errorf("%s: %s has invalid result %d signature", filePath, functionName, idx+1)
			}
		}
		return nil
	}

	return fs.ErrNotExist
}

func importAliases(file *ast.File) map[string]string {
	imports := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		importPath := strings.Trim(spec.Path.Value, `"`)
		alias := path.Base(importPath)
		if spec.Name != nil && strings.TrimSpace(spec.Name.Name) != "" {
			alias = spec.Name.Name
		}
		imports[alias] = importPath
	}
	return imports
}

func fieldTypes(fields *ast.FieldList) []ast.Expr {
	if fields == nil {
		return nil
	}
	types := make([]ast.Expr, 0, len(fields.List))
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for idx := 0; idx < count; idx++ {
			types = append(types, field.Type)
		}
	}
	return types
}

func isRuntimeContextType(expr ast.Expr, imports map[string]string, viewImport string) bool {
	indexExpr, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	if !isImportedSelector(indexExpr.X, imports, frameworkModulePath+"/framework", "RuntimeContext") {
		return false
	}
	return isPointerToImportedSelector(indexExpr.Index, imports, viewImport, "Context")
}

func isImportedSelector(expr ast.Expr, imports map[string]string, importPath string, selector string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || strings.TrimSpace(sel.Sel.Name) != selector {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return imports[strings.TrimSpace(ident.Name)] == importPath
}

func isPointerToImportedSelector(expr ast.Expr, imports map[string]string, importPath string, selector string) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	return isImportedSelector(star.X, imports, importPath, selector)
}

func isSliceOfImportedSelector(expr ast.Expr, imports map[string]string, importPath string, selector string) bool {
	arrayType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	return isImportedSelector(arrayType.Elt, imports, importPath, selector)
}

func isErrorType(expr ast.Expr, _ map[string]string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && strings.TrimSpace(ident.Name) == "error"
}

func isStringType(expr ast.Expr, _ map[string]string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && strings.TrimSpace(ident.Name) == "string"
}

func isIdentType(name string) func(ast.Expr, map[string]string) bool {
	return func(expr ast.Expr, _ map[string]string) bool {
		ident, ok := expr.(*ast.Ident)
		return ok && strings.TrimSpace(ident.Name) == strings.TrimSpace(name)
	}
}

func sortSitemapConventions(conventions []sitemapConvention) {
	sort.Slice(conventions, func(i int, j int) bool {
		return discoveryConventionLess(conventions[i].RouteID, conventions[j].RouteID)
	})
}

func sortFeedConventions(conventions []feedConvention) {
	sort.Slice(conventions, func(i int, j int) bool {
		return discoveryConventionLess(conventions[i].RouteID, conventions[j].RouteID)
	})
}

func validateMethodRoutes(paths projectlayout.ProjectLayout, routes []methodRouteDef) error {
	for idx := range routes {
		route := &routes[idx]
		methods := make([]string, 0, 7)
		for _, method := range []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		} {
			if err := validateDiscoveryFunction(
				route.SourcePath,
				method,
				[]func(ast.Expr, map[string]string) bool{
					func(expr ast.Expr, imports map[string]string) bool {
						return isRuntimeContextType(expr, imports, viewImportPath(paths))
					},
					func(expr ast.Expr, imports map[string]string) bool {
						return isImportedSelector(expr, imports, "net/http", "ResponseWriter")
					},
					func(expr ast.Expr, imports map[string]string) bool {
						return isPointerToImportedSelector(expr, imports, "net/http", "Request")
					},
					isIdentType(route.ParamsTypeName),
				},
				[]func(ast.Expr, map[string]string) bool{
					isErrorType,
				},
			); err == nil {
				methods = append(methods, method)
			} else if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
		if len(methods) == 0 {
			return fmt.Errorf("%s: route.go must declare at least one supported method function", route.SourcePath)
		}
		route.Methods = methods
	}
	return nil
}

func validateRouteConflicts(pages []templateDef, methodRoutes []methodRouteDef) error {
	seen := make(map[string]string, len(pages)+len(methodRoutes))
	for _, page := range pages {
		key := patternKeyForSegments(page.PublicSegments)
		if existing, ok := seen[key]; ok {
			return fmt.Errorf("route pattern conflict: %q and %q", existing, page.RouteID)
		}
		seen[key] = page.RouteID
	}
	for _, route := range methodRoutes {
		key := patternKeyForSegments(route.PublicSegments)
		if existing, ok := seen[key]; ok {
			return fmt.Errorf("route pattern conflict: %q and %q", existing, route.InternalRouteID)
		}
		seen[key] = route.InternalRouteID
	}
	return nil
}

func patternKeyForSegments(segments []routeSegment) string {
	if len(segments) == 0 {
		return "/"
	}
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		key := segment.PatternKeyPart()
		if key == "" {
			continue
		}
		parts = append(parts, key)
	}
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}

func buildSlotOwners(
	slotMetas []routeMeta,
	slotLayouts map[string]templateDef,
	defaults map[string]templateDef,
) map[string][]slotDef {
	byOwner := make(map[string]map[string]*slotDef)

	ensureSlot := func(ownerRouteID string, slotName string, slotRoot string) *slotDef {
		slots := byOwner[ownerRouteID]
		if slots == nil {
			slots = make(map[string]*slotDef)
			byOwner[ownerRouteID] = slots
		}
		slot := slots[slotName]
		if slot == nil {
			slot = &slotDef{
				Name:         slotName,
				RootInternal: slotRoot,
				Layouts:      make(map[string]templateDef),
			}
			slots[slotName] = slot
		}
		if slot.RootInternal == "" {
			slot.RootInternal = slotRoot
		}
		return slot
	}

	for _, meta := range slotMetas {
		slot := ensureSlot(meta.SlotOwnerRouteID, meta.SlotName, meta.SlotRootInternalID)
		slot.Pages = append(slot.Pages, meta)
	}
	for _, layout := range slotLayouts {
		slot := ensureSlot(layout.SlotOwnerRouteID, layout.SlotName, layout.SlotRootInternalID)
		slot.Layouts[layout.InternalRouteID] = layout
	}
	for slotRoot, fallback := range defaults {
		ownerRouteID, slotName, ok := slotOwnerInfo(slotRoot)
		if !ok {
			continue
		}
		slot := ensureSlot(ownerRouteID, slotName, slotRoot)
		fallbackCopy := fallback
		slot.Default = &fallbackCopy
	}

	out := make(map[string][]slotDef, len(byOwner))
	for ownerRouteID, slots := range byOwner {
		names := make([]string, 0, len(slots))
		for name := range slots {
			names = append(names, name)
		}
		sort.Strings(names)
		defs := make([]slotDef, 0, len(names))
		for _, name := range names {
			slot := *slots[name]
			sort.Slice(slot.Pages, func(i int, j int) bool {
				if slot.Pages[i].PublicRouteID != slot.Pages[j].PublicRouteID {
					return slot.Pages[i].PublicRouteID < slot.Pages[j].PublicRouteID
				}
				return slot.Pages[i].RouteID < slot.Pages[j].RouteID
			})
			defs = append(defs, slot)
		}
		out[ownerRouteID] = defs
	}
	return out
}

func slotOwnerInfo(slotRoot string) (string, string, bool) {
	segments := segmentsFromRouteID(slotRoot)
	for idx, segment := range segments {
		if !segment.IsSlot() {
			continue
		}
		return routeIDFromSegments(segments[:idx]), segment.Name, true
	}
	return "", "", false
}

func slotNamesByRoute(slotOwners map[string][]slotDef) map[string][]string {
	out := make(map[string][]string, len(slotOwners))
	for routeID, slots := range slotOwners {
		names := make([]string, 0, len(slots))
		for _, slot := range slots {
			names = append(names, slot.Name)
		}
		sort.Strings(names)
		out[routeID] = names
	}
	return out
}

func discoveryConventionLess(leftRouteID string, rightRouteID string) bool {
	leftSegments := segmentsFromRouteID(leftRouteID)
	rightSegments := segmentsFromRouteID(rightRouteID)

	leftStatic := countStaticSegments(leftSegments)
	rightStatic := countStaticSegments(rightSegments)
	if leftStatic != rightStatic {
		return leftStatic > rightStatic
	}
	if len(leftSegments) != len(rightSegments) {
		return len(leftSegments) > len(rightSegments)
	}
	return leftRouteID < rightRouteID
}

func countStaticSegments(segments []routeSegment) int {
	count := 0
	for _, segment := range segments {
		if !segment.IsParam() {
			count++
		}
	}
	return count
}

func validateDiscoveryPatternConflicts[T any](conventions []T, patternFor func(T) string) error {
	seen := make(map[string]struct{}, len(conventions))
	for _, convention := range conventions {
		pattern := discoveryPatternKey(patternFor(convention))
		if _, ok := seen[pattern]; ok {
			return fmt.Errorf("discovery route pattern conflict: %q", patternFor(convention))
		}
		seen[pattern] = struct{}{}
	}
	return nil
}

func discoveryPatternKey(pattern string) string {
	segments := segmentsFromRouteID(strings.TrimPrefix(strings.TrimSpace(pattern), "/"))
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment.IsParam() {
			parts = append(parts, ":")
			continue
		}
		parts = append(parts, segment.Name)
	}
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}
