package approutegen

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	frameworkrouter "github.com/RevoTale/no-js/framework/router"
)

func discoverRouteFiles(appRoot string, outputRoot string) (routeFiles, error) {
	templates := make([]templateDef, 0, 16)
	pages := make([]templateDef, 0, 8)
	slotPages := make([]templateDef, 0, 8)
	layouts := make(map[string]templateDef)
	slotLayouts := make(map[string]templateDef)
	defaults := make(map[string]templateDef)
	notFounds := make(map[string]templateDef)
	var rootLayoutTemplate templateDef
	discovery := discoveryConventions{}
	methodRoutes := make([]methodRouteDef, 0, 4)
	goFilesByDir := make(map[string][]string)
	usedSourceDirs := make(map[string]parsedRouteDir)
	layoutSlots := make(map[string]map[string]struct{})

	walkErr := filepath.WalkDir(appRoot, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(appRoot, filePath)
		if relErr != nil {
			return fmt.Errorf("resolve relative path for %q: %w", filePath, relErr)
		}
		relPath = filepath.ToSlash(relPath)
		routeDir := path.Dir(relPath)
		if routeDir == "." {
			routeDir = ""
		}
		dirMeta, parseErr := parseRouteDir(filepath.ToSlash(filepath.Dir(filePath)), routeDir)
		if parseErr != nil {
			return fmt.Errorf("parse route in %q: %w", relPath, parseErr)
		}

		if strings.HasSuffix(relPath, ".go") && !strings.HasSuffix(relPath, "_test.go") {
			goFilesByDir[dirMeta.SourceDir] = append(goFilesByDir[dirMeta.SourceDir], filepath.ToSlash(filePath))
		}

		if relPath == "robots.go" {
			if dirMeta.Namespace == namespaceSlot {
				return fmt.Errorf("robots.go is not allowed inside slot directories: %q", relPath)
			}
			discovery.RobotsFile = filePath
			usedSourceDirs[dirMeta.SourceDir] = dirMeta
			return nil
		}
		if path.Base(relPath) == "sitemap.go" {
			if dirMeta.Namespace == namespaceSlot {
				return fmt.Errorf("sitemap.go is not allowed inside slot directories: %q", relPath)
			}
			discovery.Sitemaps = append(discovery.Sitemaps, sitemapConvention{
				InternalRouteID: dirMeta.InternalRouteID,
				RouteID:         dirMeta.PublicRouteID,
				Segments:        slices.Clone(dirMeta.PublicSegments),
				SourcePath:      filePath,
				ImportAlias:     discoveryImportAlias(dirMeta.InternalRouteID),
				PackageModule:   sourceModuleName(dirMeta.Segments),
			})
			usedSourceDirs[dirMeta.SourceDir] = dirMeta
			return nil
		}
		if path.Base(relPath) == "feed.go" {
			if dirMeta.Namespace == namespaceSlot {
				return fmt.Errorf("feed.go is not allowed inside slot directories: %q", relPath)
			}
			discovery.Feeds = append(discovery.Feeds, feedConvention{
				InternalRouteID: dirMeta.InternalRouteID,
				RouteID:         dirMeta.PublicRouteID,
				Segments:        slices.Clone(dirMeta.PublicSegments),
				SourcePath:      filePath,
				ImportAlias:     discoveryImportAlias(dirMeta.InternalRouteID),
				PackageModule:   sourceModuleName(dirMeta.Segments),
			})
			usedSourceDirs[dirMeta.SourceDir] = dirMeta
			return nil
		}
		if path.Base(relPath) == "route.go" {
			if dirMeta.Namespace == namespaceSlot {
				return fmt.Errorf("route.go is not allowed inside slot directories: %q", relPath)
			}
			params, paramsErr := routeParamsFromSegments(dirMeta.PublicRouteID, dirMeta.PublicSegments)
			if paramsErr != nil {
				return paramsErr
			}
			methodRoutes = append(methodRoutes, methodRouteDef{
				InternalRouteID: dirMeta.InternalRouteID,
				PublicRouteID:   dirMeta.PublicRouteID,
				Segments:        slices.Clone(dirMeta.Segments),
				PublicSegments:  slices.Clone(dirMeta.PublicSegments),
				RouteName:       routeNameFromSegments(dirMeta.Segments),
				ParamsTypeName:  routeNameFromSegments(dirMeta.Segments) + "Params",
				Params:          params,
				PackageModule:   sourceModuleName(dirMeta.Segments),
				PackageAlias:    discoveryImportAlias(dirMeta.InternalRouteID),
				SourcePath:      filepath.ToSlash(filePath),
			})
			usedSourceDirs[dirMeta.SourceDir] = dirMeta
			return nil
		}

		if !strings.HasSuffix(relPath, ".templ") {
			return nil
		}

		if strings.HasPrefix(relPath, "components/") || strings.Contains(relPath, "/components/") {
			return fmt.Errorf("component templates must be under web/components: %q", relPath)
		}

		base := path.Base(relPath)
		var kind templateKind
		switch base {
		case "page.templ":
			kind = pageTemplate
		case "layout.templ":
			kind = layoutTemplate
		case "default.templ":
			kind = defaultTemplate
		case "404.templ":
			kind = notFoundTemplate
		case "error.templ":
			kind = errorTemplate
		case "root.templ":
			kind = rootTemplate
		default:
			return fmt.Errorf(
				"unsupported route template %q; only page.templ, layout.templ, default.templ, 404.templ, "+
					"and root.templ are generated; legacy error.templ files are ignored",
				relPath,
			)
		}

		if kind == rootTemplate && routeDir != "" {
			return fmt.Errorf("root.templ must be defined at web/routes/root.templ, got %q", relPath)
		}
		if kind == errorTemplate {
			return nil
		}
		if dirMeta.Namespace == namespaceSlot {
			switch kind {
			case rootTemplate, notFoundTemplate:
				return fmt.Errorf("%s is not allowed inside slot directories: %q", base, relPath)
			}
			if kind == defaultTemplate && dirMeta.InternalRouteID != dirMeta.SlotRootInternal {
				return fmt.Errorf("default.templ is only allowed at the slot root: %q", relPath)
			}
		}
		if kind == defaultTemplate && dirMeta.Namespace != namespaceSlot {
			return fmt.Errorf("default.templ is only allowed inside slot directories: %q", relPath)
		}
		moduleName := moduleNameFor(kind, dirMeta.Segments)
		tpl := templateDef{
			Kind:               kind,
			RouteID:            dirMeta.InternalRouteID,
			InternalRouteID:    dirMeta.InternalRouteID,
			PublicRouteID:      dirMeta.PublicRouteID,
			SourcePath:         filepath.ToSlash(filePath),
			Segments:           slices.Clone(dirMeta.Segments),
			PublicSegments:     slices.Clone(dirMeta.PublicSegments),
			RelativeSegments:   slices.Clone(dirMeta.RelativeSegments),
			Namespace:          dirMeta.Namespace,
			SlotName:           dirMeta.SlotName,
			SlotOwnerRouteID:   dirMeta.SlotOwnerRouteID,
			SlotRootInternalID: dirMeta.SlotRootInternal,
			ModuleName:         moduleName,
			Package:            moduleName,
			OutputDir:          filepath.ToSlash(filepath.Join(outputRoot, moduleName)),
			OutputFile:         templateOutputFileName(kind),
		}
		templates = append(templates, tpl)
		if dirMeta.Namespace == namespaceSlot && dirMeta.SlotName != "" {
			slotSet := layoutSlots[dirMeta.SlotOwnerRouteID]
			if slotSet == nil {
				slotSet = make(map[string]struct{})
				layoutSlots[dirMeta.SlotOwnerRouteID] = slotSet
			}
			slotSet[dirMeta.SlotName] = struct{}{}
		}
		if kind == pageTemplate && dirMeta.Namespace == namespaceMain {
			pages = append(pages, tpl)
		}
		if kind == pageTemplate && dirMeta.Namespace == namespaceSlot {
			slotPages = append(slotPages, tpl)
		}
		if kind == layoutTemplate {
			if dirMeta.Namespace == namespaceSlot {
				slotLayouts[dirMeta.InternalRouteID] = tpl
			} else {
				layouts[dirMeta.InternalRouteID] = tpl
			}
		}
		if kind == defaultTemplate {
			defaults[dirMeta.SlotRootInternal] = tpl
		}
		if kind == notFoundTemplate {
			notFounds[dirMeta.InternalRouteID] = tpl
		}
		if kind == rootTemplate {
			rootLayoutTemplate = tpl
		}

		return nil
	})
	if walkErr != nil {
		return routeFiles{}, fmt.Errorf("walk app templates: %w", walkErr)
	}

	sort.Slice(templates, func(i int, j int) bool {
		left := templates[i]
		right := templates[j]
		if left.PublicRouteID != right.PublicRouteID {
			return left.PublicRouteID < right.PublicRouteID
		}
		if left.RouteID != right.RouteID {
			return left.RouteID < right.RouteID
		}
		return left.Kind < right.Kind
	})
	sort.Slice(pages, func(i int, j int) bool {
		if pages[i].PublicRouteID != pages[j].PublicRouteID {
			return pages[i].PublicRouteID < pages[j].PublicRouteID
		}
		return pages[i].RouteID < pages[j].RouteID
	})
	sort.Slice(slotPages, func(i int, j int) bool {
		if slotPages[i].SlotOwnerRouteID != slotPages[j].SlotOwnerRouteID {
			return slotPages[i].SlotOwnerRouteID < slotPages[j].SlotOwnerRouteID
		}
		if slotPages[i].SlotName != slotPages[j].SlotName {
			return slotPages[i].SlotName < slotPages[j].SlotName
		}
		if slotPages[i].PublicRouteID != slotPages[j].PublicRouteID {
			return slotPages[i].PublicRouteID < slotPages[j].PublicRouteID
		}
		return slotPages[i].RouteID < slotPages[j].RouteID
	})

	if err := validateRouteConflicts(pages, methodRoutes); err != nil {
		return routeFiles{}, err
	}
	for ownerRouteID, slotSet := range layoutSlots {
		if _, ok := layouts[ownerRouteID]; !ok {
			return routeFiles{}, fmt.Errorf("slot owner %q requires a same-level layout.templ", ownerRouteID)
		}
		_ = slotSet
	}

	sourcePackages := make([]sourcePackageDef, 0, len(usedSourceDirs))
	for sourceDir, dirMeta := range usedSourceDirs {
		files := dedupeSorted(goFilesByDir[sourceDir])
		if len(files) == 0 {
			continue
		}
		params, err := routeParamsFromSegments(dirMeta.InternalRouteID, dirMeta.PublicSegments)
		if err != nil {
			return routeFiles{}, err
		}
		routeName := routeNameFromSegments(dirMeta.Segments)
		sourcePackages = append(sourcePackages, sourcePackageDef{
			InternalRouteID: dirMeta.InternalRouteID,
			PublicRouteID:   dirMeta.PublicRouteID,
			RouteName:       routeName,
			ParamsTypeName:  routeName + "Params",
			Params:          params,
			ModuleName:      sourceModuleName(dirMeta.Segments),
			Package:         sourceModuleName(dirMeta.Segments),
			SourceDir:       sourceDir,
			Files:           files,
		})
	}
	sort.Slice(sourcePackages, func(i int, j int) bool {
		return sourcePackages[i].InternalRouteID < sourcePackages[j].InternalRouteID
	})

	layoutSlotNames := make(map[string][]string, len(layoutSlots))
	for routeID, slotSet := range layoutSlots {
		names := make([]string, 0, len(slotSet))
		for name := range slotSet {
			names = append(names, name)
		}
		sort.Strings(names)
		layoutSlotNames[routeID] = names
	}

	return routeFiles{
		Templates:      templates,
		Pages:          pages,
		SlotPages:      slotPages,
		Layouts:        layouts,
		SlotLayouts:    slotLayouts,
		Defaults:       defaults,
		NotFounds:      notFounds,
		Root:           rootLayoutTemplate,
		Discovery:      discovery,
		MethodRoutes:   methodRoutes,
		SourcePackages: sourcePackages,
		LayoutSlots:    layoutSlotNames,
	}, nil
}

func parseRouteDir(sourceDir string, routeDir string) (parsedRouteDir, error) {
	rawSegments, err := frameworkrouter.ParseDirectorySegments(routeDir)
	if err != nil {
		return parsedRouteDir{}, err
	}

	segments := make([]routeSegment, 0, len(rawSegments))
	for _, raw := range rawSegments {
		segments = append(segments, newRouteSegment(raw))
	}

	publicSegments := publicSegmentsFromSegments(segments)
	parsed := parsedRouteDir{
		SourceDir:       sourceDir,
		Segments:        segments,
		PublicSegments:  publicSegments,
		InternalRouteID: routeIDFromSegments(segments),
		PublicRouteID:   publicRouteIDFromSegments(segments),
		Namespace:       namespaceMain,
	}

	slotIndex := -1
	for idx, segment := range segments {
		if !segment.IsSlot() {
			continue
		}
		if slotIndex >= 0 {
			return parsedRouteDir{}, fmt.Errorf("nested slots are not allowed in %q", routeDir)
		}
		slotIndex = idx
	}
	if slotIndex >= 0 {
		parsed.Namespace = namespaceSlot
		parsed.SlotName = segments[slotIndex].Name
		parsed.SlotOwnerRouteID = routeIDFromSegments(segments[:slotIndex])
		parsed.SlotRootInternal = routeIDFromSegments(segments[:slotIndex+1])
		parsed.RelativeSegments = publicSegmentsFromSegments(segments[slotIndex+1:])
	}

	return parsed, nil
}

func publicSegmentsFromSegments(segments []routeSegment) []routeSegment {
	out := make([]routeSegment, 0, len(segments))
	for _, segment := range segments {
		if !segment.ContributesToPublicPath() {
			continue
		}
		out = append(out, segment)
	}
	return out
}

func routeIDFromSegments(segments []routeSegment) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		parts = append(parts, segment.RoutePart())
	}
	return strings.Join(parts, "/")
}

func publicRouteIDFromSegments(segments []routeSegment) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if !segment.ContributesToPublicPath() {
			continue
		}
		parts = append(parts, segment.PublicPart())
	}
	return strings.Join(parts, "/")
}

func moduleNameFor(kind templateKind, segments []routeSegment) string {
	parts := make([]string, 0, len(segments)+2)
	parts = append(parts, "r", string(kind))
	if len(segments) == 0 {
		parts = append(parts, "root")
	} else {
		for _, segment := range segments {
			parts = append(parts, segment.SafePart())
		}
	}
	return strings.Join(parts, "_")
}

func templateOutputFileName(kind templateKind) string {
	if kind == notFoundTemplate {
		return "404.templ"
	}
	if kind == errorTemplate {
		return "error.templ"
	}
	if kind == rootTemplate {
		return "root.templ"
	}
	if kind == defaultTemplate {
		return "default.templ"
	}
	return string(kind) + ".templ"
}
