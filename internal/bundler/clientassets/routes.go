package clientassets

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	frameworkrouter "github.com/RevoTale/no-js/framework/router"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func DiscoverRouteSpecs(layout projectlayout.ProjectLayout) ([]RouteSpec, error) {
	discovery, err := discoverRouteAssets(layout)
	if err != nil {
		return nil, err
	}

	routeIDs := make([]string, 0, len(discovery.pages))
	for routeID := range discovery.pages {
		if isSlotRouteID(routeID) {
			continue
		}
		routeIDs = append(routeIDs, routeID)
	}
	sort.Strings(routeIDs)

	routes := make([]RouteSpec, 0, len(routeIDs))
	for _, routeID := range routeIDs {
		templates, cssBundles := clientAssetTemplatesForRoute(
			discovery.rootTemplate,
			routeID,
			discovery.layoutByRoute,
			discovery.slotLayoutByRoute,
			discovery.slotTemplatesByOwner,
			discovery.pages[routeID],
		)
		routes = append(routes, RouteSpec{
			RouteID:       routeID,
			TemplatePaths: templates,
			CSSBundles:    cssBundles,
		})
	}
	return routes, nil
}

func DiscoverNotFoundSpecs(layout projectlayout.ProjectLayout) ([]RouteSpec, error) {
	discovery, err := discoverRouteAssets(layout)
	if err != nil {
		return nil, err
	}

	routeIDs := make([]string, 0, len(discovery.notFounds))
	for routeID := range discovery.notFounds {
		if isSlotRouteID(routeID) {
			continue
		}
		routeIDs = append(routeIDs, routeID)
	}
	sort.Strings(routeIDs)

	routes := make([]RouteSpec, 0, len(routeIDs))
	for _, routeID := range routeIDs {
		templates, cssBundles := clientAssetTemplatesForRoute(
			discovery.rootTemplate,
			routeID,
			discovery.layoutByRoute,
			discovery.slotLayoutByRoute,
			discovery.slotTemplatesByOwner,
			discovery.notFounds[routeID],
		)
		routes = append(routes, RouteSpec{
			RouteID:       routeID,
			TemplatePaths: templates,
			CSSBundles:    cssBundles,
		})
	}
	return routes, nil
}

type routeAssetDiscovery struct {
	rootTemplate         string
	layoutByRoute        map[string]string
	slotLayoutByRoute    map[string]string
	pages                map[string]string
	notFounds            map[string]string
	slotTemplatesByOwner map[string][]slotTemplateRef
}

type routeTemplateRef struct {
	RouteID    string
	SourcePath string
}

type slotTemplateRef struct {
	RouteID         string
	SlotRootRouteID string
	SourcePath      string
	Name            string
}

func discoverRouteAssets(layout projectlayout.ProjectLayout) (*routeAssetDiscovery, error) {
	root := strings.TrimSpace(layout.RoutesDir)
	if root == "" {
		return nil, fmt.Errorf("routes dir is required")
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return &routeAssetDiscovery{
				layoutByRoute:        map[string]string{},
				slotLayoutByRoute:    map[string]string{},
				pages:                map[string]string{},
				notFounds:            map[string]string{},
				slotTemplatesByOwner: map[string][]slotTemplateRef{},
			}, nil
		}
		return nil, err
	}

	discovery := &routeAssetDiscovery{
		layoutByRoute:        map[string]string{},
		slotLayoutByRoute:    map[string]string{},
		pages:                map[string]string{},
		notFounds:            map[string]string{},
		slotTemplatesByOwner: map[string][]slotTemplateRef{},
	}
	if err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		switch name {
		case "root.templ", "layout.templ", "page.templ", "default.templ", "404.templ":
		default:
			return nil
		}

		dir := filepath.Dir(filePath)
		routeID, err := routeIDForDir(root, dir)
		if err != nil {
			return err
		}
		ownerRouteID, slotRootRouteID, isSlot, err := slotRouteInfo(root, dir)
		if err != nil {
			return err
		}
		if isSlot {
			discovery.slotTemplatesByOwner[ownerRouteID] = append(
				discovery.slotTemplatesByOwner[ownerRouteID],
				slotTemplateRef{
					RouteID:         routeID,
					SlotRootRouteID: slotRootRouteID,
					SourcePath:      filePath,
					Name:            name,
				},
			)
		}

		switch name {
		case "root.templ":
			discovery.rootTemplate = filePath
		case "layout.templ":
			if isSlot {
				discovery.slotLayoutByRoute[routeID] = filePath
			} else {
				discovery.layoutByRoute[routeID] = filePath
			}
		case "page.templ":
			discovery.pages[routeID] = filePath
		case "404.templ":
			discovery.notFounds[routeID] = filePath
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for ownerRouteID := range discovery.slotTemplatesByOwner {
		sort.Slice(discovery.slotTemplatesByOwner[ownerRouteID], func(i int, j int) bool {
			return discovery.slotTemplatesByOwner[ownerRouteID][i].SourcePath <
				discovery.slotTemplatesByOwner[ownerRouteID][j].SourcePath
		})
	}
	return discovery, nil
}

func clientAssetTemplatesForRoute(
	rootTemplate string,
	routeID string,
	layoutByRoute map[string]string,
	slotLayoutByRoute map[string]string,
	slotTemplatesByOwner map[string][]slotTemplateRef,
	terminalTemplate string,
) ([]string, []CSSBundleSpec) {
	templates := []string{}
	seenTemplates := map[string]struct{}{}
	cssBundles := []CSSBundleSpec{}

	if strings.TrimSpace(rootTemplate) != "" {
		appendUniquePath(&templates, seenTemplates, rootTemplate)
		cssBundles = append(cssBundles, CSSBundleSpec{
			OwnerTemplatePath: rootTemplate,
			TemplatePaths:     []string{rootTemplate},
		})
	}

	layoutRefs := layoutTemplateChain(routeID, layoutByRoute)
	lastNonRootLayout := -1
	for _, layout := range layoutRefs {
		appendUniquePath(&templates, seenTemplates, layout.SourcePath)
		bundlePaths := []string{layout.SourcePath}
		for _, slotTemplate := range clientAssetSlotTemplatesForRoute(
			slotTemplatesByOwner[layout.RouteID],
			slotLayoutByRoute,
		) {
			appendUniquePath(&templates, seenTemplates, slotTemplate)
			bundlePaths = append(bundlePaths, slotTemplate)
		}
		if layout.RouteID != "" {
			lastNonRootLayout = len(cssBundles)
		}
		cssBundles = append(cssBundles, CSSBundleSpec{
			OwnerTemplatePath: layout.SourcePath,
			TemplatePaths:     bundlePaths,
		})
	}

	if strings.TrimSpace(terminalTemplate) != "" {
		appendUniquePath(&templates, seenTemplates, terminalTemplate)
		if lastNonRootLayout >= 0 {
			cssBundles[lastNonRootLayout].TemplatePaths = append(
				cssBundles[lastNonRootLayout].TemplatePaths,
				terminalTemplate,
			)
		} else {
			cssBundles = append(cssBundles, CSSBundleSpec{
				OwnerTemplatePath: terminalTemplate,
				TemplatePaths:     []string{terminalTemplate},
			})
		}
	}

	return templates, cssBundles
}

func clientAssetSlotTemplatesForRoute(
	refs []slotTemplateRef,
	slotLayoutByRoute map[string]string,
) []string {
	if len(refs) == 0 {
		return nil
	}

	bySlotRoot := make(map[string][]slotTemplateRef)
	for _, ref := range refs {
		if strings.TrimSpace(ref.SlotRootRouteID) == "" {
			continue
		}
		bySlotRoot[ref.SlotRootRouteID] = append(bySlotRoot[ref.SlotRootRouteID], ref)
	}

	roots := make([]string, 0, len(bySlotRoot))
	for root := range bySlotRoot {
		roots = append(roots, root)
	}
	sort.Strings(roots)

	templates := []string{}
	seen := map[string]struct{}{}
	for _, root := range roots {
		group := bySlotRoot[root]
		appendUniqueTemplateRefs(&templates, seen, layoutTemplateChain(root, slotLayoutByRoute)...)
		if fallback, ok := slotTemplateByName(group, "default.templ"); ok {
			appendUniquePath(&templates, seen, fallback.SourcePath)
		}
		for _, page := range slotTemplatesByName(group, "page.templ") {
			appendUniqueTemplateRefs(&templates, seen, layoutTemplateChain(page.RouteID, slotLayoutByRoute)...)
			appendUniquePath(&templates, seen, page.SourcePath)
		}
	}
	return templates
}

func slotTemplatesByName(refs []slotTemplateRef, name string) []slotTemplateRef {
	matches := make([]slotTemplateRef, 0, len(refs))
	for _, ref := range refs {
		if ref.Name == name {
			matches = append(matches, ref)
		}
	}
	sort.Slice(matches, func(i int, j int) bool {
		if matches[i].RouteID != matches[j].RouteID {
			return matches[i].RouteID < matches[j].RouteID
		}
		return matches[i].SourcePath < matches[j].SourcePath
	})
	return matches
}

func slotTemplateByName(refs []slotTemplateRef, name string) (slotTemplateRef, bool) {
	for _, ref := range refs {
		if ref.Name == name {
			return ref, true
		}
	}
	return slotTemplateRef{}, false
}

func appendUniqueTemplateRefs(paths *[]string, seen map[string]struct{}, refs ...routeTemplateRef) {
	for _, ref := range refs {
		appendUniquePath(paths, seen, ref.SourcePath)
	}
}

func appendUniquePath(paths *[]string, seen map[string]struct{}, filePath string) {
	if strings.TrimSpace(filePath) == "" {
		return
	}
	if _, ok := seen[filePath]; ok {
		return
	}
	seen[filePath] = struct{}{}
	*paths = append(*paths, filePath)
}

func layoutTemplateChain(routeID string, layoutByRoute map[string]string) []routeTemplateRef {
	segments := []string{}
	if strings.TrimSpace(routeID) != "" {
		segments = strings.Split(routeID, "/")
	}
	candidates := []string{""}
	for idx := 1; idx <= len(segments); idx++ {
		candidates = append(candidates, strings.Join(segments[:idx], "/"))
	}
	refs := make([]routeTemplateRef, 0, len(candidates))
	for _, candidate := range candidates {
		if filePath := layoutByRoute[candidate]; strings.TrimSpace(filePath) != "" {
			refs = append(refs, routeTemplateRef{RouteID: candidate, SourcePath: filePath})
		}
	}
	return refs
}

func slotRouteInfo(root string, dir string) (string, string, bool, error) {
	relative, err := filepath.Rel(root, dir)
	if err != nil {
		return "", "", false, err
	}
	relative = filepath.ToSlash(relative)
	if relative == "." {
		return "", "", false, nil
	}
	segments := strings.Split(relative, "/")
	for index, segment := range segments {
		if !strings.HasPrefix(segment, "_slot__") {
			continue
		}
		ownerDir := root
		if index > 0 {
			ownerDir = filepath.Join(root, filepath.FromSlash(strings.Join(segments[:index], "/")))
		}
		ownerRouteID, err := routeIDForDir(root, ownerDir)
		if err != nil {
			return "", "", false, err
		}
		slotRootDir := filepath.Join(root, filepath.FromSlash(strings.Join(segments[:index+1], "/")))
		slotRootRouteID, err := routeIDForDir(root, slotRootDir)
		if err != nil {
			return "", "", false, err
		}
		return ownerRouteID, slotRootRouteID, true, nil
	}
	return "", "", false, nil
}

func isSlotRouteID(routeID string) bool {
	for _, segment := range strings.Split(routeID, "/") {
		if strings.HasPrefix(segment, "_slot__") {
			return true
		}
	}
	return false
}

func routeIDForDir(root string, dir string) (string, error) {
	relative, err := filepath.Rel(root, dir)
	if err != nil {
		return "", err
	}
	relative = filepath.ToSlash(relative)
	if relative == "." {
		relative = ""
	}
	rawSegments, err := frameworkrouter.ParseDirectorySegments(relative)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		parts = append(parts, segment.RawPart())
	}
	return strings.Join(parts, "/"), nil
}
