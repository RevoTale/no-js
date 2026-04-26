package appshape

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/clientassetext"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

var (
	routeTemplateNames = map[string]struct{}{
		"root.templ":    {},
		"page.templ":    {},
		"layout.templ":  {},
		"default.templ": {},
		"404.templ":     {},
	}
	routeGoNames = map[string]struct{}{
		"route.go":   {},
		"robots.go":  {},
		"feed.go":    {},
		"sitemap.go": {},
	}
	routeAssetStems = map[string]struct{}{
		"page":   {},
		"layout": {},
		"404":    {},
	}
)

// Validate checks the strict app-owned source directories that feed generation.
func Validate(layout projectlayout.ProjectLayout) error {
	if err := validateRoutes(layout.RoutesDir); err != nil {
		return err
	}
	if err := validateComponents(componentsDir(layout)); err != nil {
		return err
	}
	return nil
}

func validateRoutes(root string) error {
	if strings.TrimSpace(root) == "" {
		return nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat route tree %q: %w", filepath.ToSlash(root), err)
	}

	routeDirs := map[string]*routeDir{}
	walkErr := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, filePath)
		if err != nil {
			return fmt.Errorf("resolve route file %q: %w", filepath.ToSlash(filePath), err)
		}
		relPath = filepath.ToSlash(relPath)
		if !isAllowedRouteFile(relPath) {
			return fmt.Errorf(
				"unsupported file in web/routes: %q; route directories may only contain route templates, "+
					"route convention Go files, and same-stem page/layout/404 Client Assets",
				relPath,
			)
		}

		dir := path.Dir(relPath)
		if dir == "." {
			dir = ""
		}
		meta := routeDirs[dir]
		if meta == nil {
			meta = &routeDir{templates: map[string]struct{}{}, assets: map[string][]string{}, scriptSources: map[string]string{}}
			routeDirs[dir] = meta
		}
		base := path.Base(relPath)
		if _, ok := routeTemplateNames[base]; ok {
			meta.templates[base] = struct{}{}
		}
		if stem, ok := routeClientAssetStem(base); ok {
			meta.assets[stem] = append(meta.assets[stem], relPath)
		}
		if stem, ok := routeClientScriptStem(base); ok {
			if meta.scriptSources[stem] != "" {
				return fmt.Errorf(
					"route directory %s has multiple script sources for %q: %q and %q; choose one of %s "+
						"because they all emit %q",
					displayRouteDir(dir),
					routeTemplatePath(dir, stem),
					meta.scriptSources[stem],
					relPath,
					clientassetext.ScriptChoices(stem),
					clientassetext.ScriptOutputName(stem),
				)
			}
			meta.scriptSources[stem] = relPath
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	dirs := make([]string, 0, len(routeDirs))
	for dir := range routeDirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		meta := routeDirs[dir]
		stems := make([]string, 0, len(meta.assets))
		for stem := range meta.assets {
			stems = append(stems, stem)
		}
		sort.Strings(stems)
		for _, stem := range stems {
			templateName := stem + ".templ"
			if _, ok := meta.templates[templateName]; ok {
				continue
			}
			assetPath := meta.assets[stem][0]
			requiredPath := templateName
			if dir != "" {
				requiredPath = dir + "/" + templateName
			}
			return fmt.Errorf(
				"route Client Asset %q requires matching template %q in the same directory because "+
					"route assets are attached to generated page, layout, or 404 endpoints",
				assetPath,
				requiredPath,
			)
		}
	}

	return nil
}

type routeDir struct {
	templates     map[string]struct{}
	assets        map[string][]string
	scriptSources map[string]string
}

func isAllowedRouteFile(relPath string) bool {
	base := path.Base(relPath)
	dir := path.Dir(relPath)
	if dir == "." {
		dir = ""
	}

	if _, ok := routeTemplateNames[base]; ok {
		return true
	}
	if isRouteClientAsset(base) || isRouteClientAssetHelper(base) {
		return true
	}
	if base == "robots.go" {
		return dir == ""
	}
	if _, ok := routeGoNames[base]; ok {
		return true
	}
	return false
}

func isRouteClientAsset(base string) bool {
	_, ok := routeClientAssetStem(base)
	return ok
}

func routeClientAssetStem(base string) (string, bool) {
	extension := path.Ext(base)
	if !clientassetext.IsAssetExtension(extension) {
		return "", false
	}
	stem := strings.TrimSuffix(base, extension)
	_, ok := routeAssetStems[stem]
	return stem, ok
}

func routeClientScriptStem(base string) (string, bool) {
	extension := path.Ext(base)
	if !clientassetext.IsScriptExtension(extension) {
		return "", false
	}
	stem := strings.TrimSuffix(base, extension)
	_, ok := routeAssetStems[stem]
	return stem, ok
}

func isRouteClientAssetHelper(base string) bool {
	stem, ok := clientassetext.GeneratedHelperStem(base)
	if !ok {
		return false
	}
	_, ok = routeAssetStems[stem]
	return ok
}

func displayRouteDir(dir string) string {
	if dir == "" {
		return "web/routes"
	}
	return "web/routes/" + dir
}

func routeTemplatePath(dir string, stem string) string {
	templatePath := stem + ".templ"
	if dir == "" {
		return templatePath
	}
	return dir + "/" + templatePath
}

func validateComponents(root string) error {
	if strings.TrimSpace(root) == "" {
		return nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat component tree %q: %w", filepath.ToSlash(root), err)
	}

	dirs := map[string]*componentDir{}
	if err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, filePath)
		if err != nil {
			return fmt.Errorf("resolve component file %q: %w", filepath.ToSlash(filePath), err)
		}
		relPath = filepath.ToSlash(relPath)
		base := path.Base(relPath)
		dir := path.Dir(relPath)
		if dir == "." {
			return fmt.Errorf(
				"unsupported file in web/components: %q; files must live in component package directories "+
					"like web/components/%s/%s",
				relPath,
				strings.TrimSuffix(base, path.Ext(base)),
				base,
			)
		}

		name := path.Base(dir)
		if err := validateComponentFile(relPath, filePath, name); err != nil {
			return err
		}

		meta := dirs[dir]
		if meta == nil {
			meta = &componentDir{}
			dirs[dir] = meta
		}
		if base == name+".templ" || base == name+".go" {
			meta.hasAnchor = true
		}
		if isComponentScriptSource(name, base) {
			if meta.scriptSource != "" {
				return fmt.Errorf(
					"component package web/components/%s has multiple script sources %q and %q; choose one of "+
						"%s because all component scripts emit %q",
					dir,
					meta.scriptSource,
					relPath,
					clientassetext.ScriptChoices(name),
					clientassetext.ScriptOutputName(name),
				)
			}
			meta.scriptSource = relPath
		}
		return nil
	}); err != nil {
		return err
	}

	names := make([]string, 0, len(dirs))
	for dir := range dirs {
		names = append(names, dir)
	}
	sort.Strings(names)
	for _, dir := range names {
		meta := dirs[dir]
		if meta.hasAnchor {
			continue
		}
		name := path.Base(dir)
		return fmt.Errorf(
			"component package web/components/%s must contain %q or %q because the directory name defines "+
				"the component package anchor",
			dir,
			name+".templ",
			name+".go",
		)
	}

	return nil
}

type componentDir struct {
	hasAnchor    bool
	scriptSource string
}

func validateComponentFile(relPath string, filePath string, name string) error {
	base := path.Base(relPath)
	extension := path.Ext(base)

	if !isValidGoPackageName(name) {
		return fmt.Errorf(
			"component package web/components/%s has invalid Go package name %q; directory names must be valid Go package names",
			path.Dir(relPath),
			name,
		)
	}

	if base == "templ_css_exports_gen.go" || base == name+"_templ.go" || clientassetext.IsGeneratedHelperFor(name, base) ||
		strings.HasSuffix(base, "_test.go") {
		return nil
	}

	switch {
	case extension == ".templ":
		if base != name+".templ" {
			return fmt.Errorf(
				"unsupported component template %q; component package web/components/%s may only declare "+
					"its public template anchor in %q. Split large markup into another component package",
				relPath,
				path.Dir(relPath),
				name+".templ",
			)
		}
		return validateTemplPackageName(filePath, relPath, name)
	case extension == ".go":
		if err := validateGoPackageName(filePath, relPath, name); err != nil {
			return err
		}
		if base == name+".go" {
			return nil
		}
		return validatePrivateComponentGoFile(filePath, relPath, name)
	case isAssetExtension(extension):
		if base != name+extension {
			return fmt.Errorf(
				"unsupported component asset %q; component Client Assets must use the component anchor stem, "+
					"for example web/components/%s/%s",
				relPath,
				path.Dir(relPath),
				name+extension,
			)
		}
		return nil
	}

	return fmt.Errorf(
		"unsupported file in web/components: %q; component packages may only contain %s, support .go files, "+
			"same-stem Client Assets, tests, and generated helpers",
		relPath,
		name+".templ or "+name+".go",
	)
}

func isAssetExtension(extension string) bool {
	return clientassetext.IsAssetExtension(extension)
}

func isComponentScriptSource(name string, base string) bool {
	extension := path.Ext(base)
	if !isScriptAssetExtension(extension) {
		return false
	}
	return base == name+extension
}

func isScriptAssetExtension(extension string) bool {
	return clientassetext.IsScriptExtension(extension)
}

func validateTemplPackageName(filePath string, relPath string, expected string) error {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read component template %q: %w", relPath, err)
	}
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "package" {
			if fields[1] == expected {
				return nil
			}
			return fmt.Errorf(
				"component template %q declares package %q; files under web/components/%s must use package %s",
				relPath,
				fields[1],
				path.Dir(relPath),
				expected,
			)
		}
		break
	}
	return fmt.Errorf("component template %q must declare package %s", relPath, expected)
}

func validateGoPackageName(filePath string, relPath string, expected string) error {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filePath, nil, parser.PackageClauseOnly)
	if err != nil {
		return fmt.Errorf("parse component Go file %q: %w", relPath, err)
	}
	if file.Name.Name == expected {
		return nil
	}
	return fmt.Errorf(
		"component Go file %q declares package %q; files under web/components/%s must use package %s",
		relPath,
		file.Name.Name,
		path.Dir(relPath),
		expected,
	)
}

func validatePrivateComponentGoFile(filePath string, relPath string, name string) error {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filePath, nil, 0)
	if err != nil {
		return fmt.Errorf("parse component Go file %q: %w", relPath, err)
	}
	for _, decl := range file.Decls {
		switch typedDecl := decl.(type) {
		case *ast.FuncDecl:
			if ast.IsExported(typedDecl.Name.Name) {
				return exportedComponentSupportError(relPath, typedDecl.Name.Name, name)
			}
		case *ast.GenDecl:
			for _, spec := range typedDecl.Specs {
				switch typedSpec := spec.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(typedSpec.Name.Name) {
						return exportedComponentSupportError(relPath, typedSpec.Name.Name, name)
					}
				case *ast.ValueSpec:
					for _, ident := range typedSpec.Names {
						if ast.IsExported(ident.Name) {
							return exportedComponentSupportError(relPath, ident.Name, name)
						}
					}
				}
			}
		}
	}
	return nil
}

func exportedComponentSupportError(relPath string, symbol string, name string) error {
	return fmt.Errorf(
		"exported declaration %q in web/components/%s must move to %q or be made private; "+
			"support Go files in component packages are private implementation",
		symbol,
		relPath,
		name+".go",
	)
}

func isValidGoPackageName(name string) bool {
	if name == "" || token.Lookup(name).IsKeyword() {
		return false
	}
	for index, value := range name {
		if value == '_' {
			continue
		}
		if index == 0 && value >= '0' && value <= '9' {
			return false
		}
		if (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9') {
			continue
		}
		return false
	}
	return true
}

func componentsDir(layout projectlayout.ProjectLayout) string {
	if strings.TrimSpace(layout.RoutesDir) == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(layout.RoutesDir), "components")
}
