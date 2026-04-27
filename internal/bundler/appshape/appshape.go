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

	frameworkrouter "github.com/RevoTale/no-js/framework/router"
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
)

type routeAssetOwner string

const (
	routeAssetOwnerRoot    routeAssetOwner = "root"
	routeAssetOwnerPage    routeAssetOwner = "page"
	routeAssetOwnerLayout  routeAssetOwner = "layout"
	routeAssetOwnerDefault routeAssetOwner = "default"
	routeAssetOwner404     routeAssetOwner = "404"
)

var routeAssetOwners = map[string]routeAssetOwner{
	"root":    routeAssetOwnerRoot,
	"page":    routeAssetOwnerPage,
	"layout":  routeAssetOwnerLayout,
	"default": routeAssetOwnerDefault,
	"404":     routeAssetOwner404,
}

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

	shape := newRouteShape()
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
		routeFile, err := newRouteFile(relPath)
		if err != nil {
			return err
		}
		if err := validateRouteFile(routeFile); err != nil {
			return err
		}

		return shape.addFile(routeFile)
	})
	if walkErr != nil {
		return walkErr
	}

	return shape.validate()
}

type routeDir struct {
	templates     map[string]struct{}
	assets        map[string][]string
	scriptSources map[string]string
}

type routeShape struct {
	dirs             map[string]*routeDir
	endpointPatterns map[string]string
	slotOwnerFiles   map[string]string
}

type routeFile struct {
	RelPath  string
	Dir      string
	Base     string
	Segments []frameworkrouter.Segment
}

func newRouteShape() *routeShape {
	return &routeShape{
		dirs:             map[string]*routeDir{},
		endpointPatterns: map[string]string{},
		slotOwnerFiles:   map[string]string{},
	}
}

func (shape *routeShape) addFile(file routeFile) error {
	meta := shape.ensureDir(file.Dir)
	if _, ok := routeTemplateNames[file.Base]; ok {
		meta.templates[file.Base] = struct{}{}
	}
	if stem, ok := routeClientAssetStem(file.Base); ok {
		meta.assets[stem] = append(meta.assets[stem], file.RelPath)
	}
	if stem, ok := routeClientScriptStem(file.Base); ok {
		if meta.scriptSources[stem] != "" {
			return fmt.Errorf(
				"route directory %s has multiple script sources for %q: %q and %q; choose one of %s "+
					"because they all emit %q",
				displayRouteDir(file.Dir),
				routeTemplatePath(file.Dir, stem),
				meta.scriptSources[stem],
				file.RelPath,
				clientassetext.ScriptChoices(stem),
				clientassetext.ScriptOutputName(stem),
			)
		}
		meta.scriptSources[stem] = file.RelPath
	}
	if ownerDir, ok := file.slotOwnerDir(); ok {
		if shape.slotOwnerFiles[ownerDir] == "" {
			shape.slotOwnerFiles[ownerDir] = file.RelPath
		}
	}
	if file.definesMainEndpoint() {
		return shape.addEndpoint(file)
	}
	return nil
}

func (shape *routeShape) ensureDir(dir string) *routeDir {
	meta := shape.dirs[dir]
	if meta == nil {
		meta = &routeDir{
			templates:     map[string]struct{}{},
			assets:        map[string][]string{},
			scriptSources: map[string]string{},
		}
		shape.dirs[dir] = meta
	}
	return meta
}

func (shape *routeShape) addEndpoint(file routeFile) error {
	key := routePatternKey(file.Segments)
	if existing := shape.endpointPatterns[key]; existing != "" {
		return fmt.Errorf(
			"route pattern conflict: %q and %q both resolve to %q; do not define page.templ and route.go "+
				"for the same public route",
			existing,
			file.RelPath,
			key,
		)
	}
	shape.endpointPatterns[key] = file.RelPath
	return nil
}

func (shape *routeShape) validate() error {
	if err := shape.validateMatchingRouteAssets(); err != nil {
		return err
	}
	return shape.validateSlotOwnerLayouts()
}

func (shape *routeShape) validateMatchingRouteAssets() error {
	dirs := make([]string, 0, len(shape.dirs))
	for dir := range shape.dirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		meta := shape.dirs[dir]
		stems := make([]string, 0, len(meta.assets))
		for stem := range meta.assets {
			stems = append(stems, stem)
		}
		sort.Strings(stems)
		for _, stem := range stems {
			templateName := stem + ".templ"
			if meta.hasTemplate(templateName) {
				continue
			}
			assetPath := meta.assets[stem][0]
			requiredPath := templateName
			if dir != "" {
				requiredPath = dir + "/" + templateName
			}
			return fmt.Errorf(
				"route Client Asset %q requires matching template %q in the same directory because "+
					"route assets are attached to generated root, page, layout, slot fallback, or 404 endpoints",
				assetPath,
				requiredPath,
			)
		}
	}
	return nil
}

func (shape *routeShape) validateSlotOwnerLayouts() error {
	ownerDirs := make([]string, 0, len(shape.slotOwnerFiles))
	for ownerDir := range shape.slotOwnerFiles {
		ownerDirs = append(ownerDirs, ownerDir)
	}
	sort.Strings(ownerDirs)
	for _, ownerDir := range ownerDirs {
		meta := shape.dirs[ownerDir]
		if meta != nil && meta.hasTemplate("layout.templ") {
			continue
		}
		return fmt.Errorf(
			"slot route file %q requires owner layout %q because slot content can only be rendered by "+
				"a same-level layout.templ",
			shape.slotOwnerFiles[ownerDir],
			routeTemplatePath(ownerDir, "layout"),
		)
	}
	return nil
}

func (meta *routeDir) hasTemplate(name string) bool {
	if meta == nil {
		return false
	}
	_, ok := meta.templates[name]
	return ok
}

func newRouteFile(relPath string) (routeFile, error) {
	dir := path.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	segments, err := frameworkrouter.ParseDirectorySegments(dir)
	if err != nil {
		return routeFile{}, fmt.Errorf("parse route directory for %q: %w", relPath, err)
	}
	return routeFile{
		RelPath:  relPath,
		Dir:      dir,
		Base:     path.Base(relPath),
		Segments: segments,
	}, nil
}

func validateRouteFile(file routeFile) error {
	if file.slotCount() > 1 {
		return fmt.Errorf("nested slots are not allowed in %q", file.RelPath)
	}
	if _, ok := routeTemplateNames[file.Base]; ok {
		return validateRouteTemplatePlacement(file)
	}
	if isRouteClientAsset(file.Base) || isRouteClientAssetHelper(file.Base) {
		return validateRouteClientAssetPlacement(file)
	}
	if _, ok := routeGoNames[file.Base]; ok {
		return validateRouteGoPlacement(file)
	}
	return unsupportedRouteFileError(file.RelPath)
}

func validateRouteTemplatePlacement(file routeFile) error {
	switch file.Base {
	case "root.templ":
		if file.Dir != "" {
			return fmt.Errorf("root.templ must be defined at web/routes/root.templ, got %q", file.RelPath)
		}
	case "default.templ":
		if !file.inSlot() {
			return fmt.Errorf("default.templ is only allowed inside slot directories: %q", file.RelPath)
		}
		if !file.isSlotRoot() {
			return fmt.Errorf("default.templ is only allowed at the slot root: %q", file.RelPath)
		}
	case "404.templ":
		if file.inSlot() {
			return fmt.Errorf("404.templ is not allowed inside slot directories: %q", file.RelPath)
		}
	}
	return nil
}

func validateRouteClientAssetPlacement(file routeFile) error {
	owner, ok := routeClientAssetOwner(file.Base)
	if !ok {
		owner, _ = routeClientAssetHelperOwner(file.Base)
	}
	if owner == "" {
		return unsupportedRouteFileError(file.RelPath)
	}
	switch owner {
	case routeAssetOwnerRoot:
		if file.Dir != "" {
			return fmt.Errorf("root.css is only allowed at web/routes/root.css, got %q", file.RelPath)
		}
	case routeAssetOwnerDefault:
		if !file.inSlot() {
			return fmt.Errorf("default Client Assets are only allowed inside slot directories: %q", file.RelPath)
		}
		if !file.isSlotRoot() {
			return fmt.Errorf("default Client Assets are only allowed at the slot root: %q", file.RelPath)
		}
	case routeAssetOwner404:
		if file.inSlot() {
			return fmt.Errorf("404 Client Assets are not allowed inside slot directories: %q", file.RelPath)
		}
	}
	return nil
}

func validateRouteGoPlacement(file routeFile) error {
	if file.inSlot() {
		return fmt.Errorf("%s is not allowed inside slot directories: %q", file.Base, file.RelPath)
	}
	if file.Base == "robots.go" && file.Dir != "" {
		return fmt.Errorf("robots.go is only allowed at web/routes/robots.go, got %q", file.RelPath)
	}
	return nil
}

func unsupportedRouteFileError(relPath string) error {
	return fmt.Errorf(
		"unsupported file in web/routes: %q; route directories may only contain route templates, "+
			"route convention Go files, and same-stem root/page/layout/default/404 Client Assets",
		relPath,
	)
}

func (file routeFile) inSlot() bool {
	return file.slotCount() > 0
}

func (file routeFile) slotCount() int {
	count := 0
	for _, segment := range file.Segments {
		if segment.Kind == frameworkrouter.SegmentSlot {
			count++
		}
	}
	return count
}

func (file routeFile) isSlotRoot() bool {
	if len(file.Segments) == 0 {
		return false
	}
	return file.Segments[len(file.Segments)-1].Kind == frameworkrouter.SegmentSlot
}

func (file routeFile) slotOwnerDir() (string, bool) {
	for idx, segment := range file.Segments {
		if segment.Kind != frameworkrouter.SegmentSlot {
			continue
		}
		if idx == 0 {
			return "", true
		}
		parts := strings.Split(file.Dir, "/")
		return strings.Join(parts[:idx], "/"), true
	}
	return "", false
}

func (file routeFile) definesMainEndpoint() bool {
	if file.inSlot() {
		return false
	}
	return file.Base == "page.templ" || file.Base == "route.go"
}

func routePatternKey(segments []frameworkrouter.Segment) string {
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		part := segment.PatternKeyPart()
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}

func isRouteClientAsset(base string) bool {
	_, ok := routeClientAssetStem(base)
	return ok
}

func routeClientAssetStem(base string) (string, bool) {
	owner, ok := routeClientAssetOwner(base)
	return string(owner), ok
}

func routeClientAssetOwner(base string) (routeAssetOwner, bool) {
	extension := path.Ext(base)
	if !clientassetext.IsAssetExtension(extension) {
		return "", false
	}
	stem := strings.TrimSuffix(base, extension)
	owner, ok := routeAssetOwners[stem]
	if !ok {
		return "", false
	}
	if owner == routeAssetOwnerRoot && extension != clientassetext.CSSExtension {
		return "", false
	}
	return owner, true
}

func routeClientScriptStem(base string) (string, bool) {
	extension := path.Ext(base)
	if !clientassetext.IsScriptExtension(extension) {
		return "", false
	}
	stem := strings.TrimSuffix(base, extension)
	owner, ok := routeAssetOwners[stem]
	if !ok || owner == routeAssetOwnerRoot {
		return "", false
	}
	return string(owner), true
}

func isRouteClientAssetHelper(base string) bool {
	_, ok := routeClientAssetHelperStem(base)
	return ok
}

func routeClientAssetHelperStem(base string) (string, bool) {
	owner, ok := routeClientAssetHelperOwner(base)
	return string(owner), ok
}

func routeClientAssetHelperOwner(base string) (routeAssetOwner, bool) {
	stem, ok := clientassetext.GeneratedHelperStem(base)
	if !ok {
		return "", false
	}
	owner, ok := routeAssetOwners[stem]
	if !ok {
		return "", false
	}
	if owner == routeAssetOwnerRoot && base != "root.css_gen.go" {
		return "", false
	}
	return owner, true
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
