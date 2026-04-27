package approutegen

import (
	"bytes"
	"sort"
	"strings"

	"github.com/RevoTale/no-js/internal/bundler/clientassets"
)

func writeContractParamsStruct(buffer *bytes.Buffer, contract routeContractDef) {
	writef(buffer, "type %s struct {\n", contract.ParamsTypeName)
	if len(contract.Params) == 0 {
		buffer.WriteString("}\n\n")
		return
	}
	for _, param := range contract.Params {
		writef(buffer, "\t%s %s\n", param.FieldName, param.Type)
	}
	buffer.WriteString("}\n\n")
}

func buildRouteContracts(
	metas []routeMeta,
	layouts map[string]templateDef,
	defaults map[string]templateDef,
	notFounds map[string]templateDef,
) ([]routeContractDef, error) {
	contractsByRoute := make(map[string]routeContractDef, len(metas)+len(layouts)+len(defaults)+len(notFounds)+1)

	for _, meta := range metas {
		contractsByRoute[meta.RouteID] = routeContractDef{
			RouteID:         meta.RouteID,
			InternalRouteID: meta.InternalRouteID,
			PublicRouteID:   meta.PublicRouteID,
			RouteName:       meta.RouteName,
			ParamsTypeName:  meta.ParamsTypeName,
			Params:          meta.Params,
		}
	}

	for routeID, layout := range layouts {
		if _, ok := contractsByRoute[routeID]; ok {
			continue
		}
		params, err := routeParamsFromSegments(layout.PublicRouteID, layout.PublicSegments)
		if err != nil {
			return nil, err
		}
		routeName := routeNameFromSegments(layout.Segments)
		contractsByRoute[routeID] = routeContractDef{
			RouteID:         routeID,
			InternalRouteID: layout.InternalRouteID,
			PublicRouteID:   layout.PublicRouteID,
			RouteName:       routeName,
			ParamsTypeName:  routeName + "Params",
			Params:          params,
		}
	}

	for routeID, fallback := range defaults {
		if _, ok := contractsByRoute[routeID]; ok {
			continue
		}
		params, err := routeParamsFromSegments(fallback.PublicRouteID, fallback.PublicSegments)
		if err != nil {
			return nil, err
		}
		routeName := routeNameFromSegments(fallback.Segments)
		contractsByRoute[routeID] = routeContractDef{
			RouteID:         routeID,
			InternalRouteID: fallback.InternalRouteID,
			PublicRouteID:   fallback.PublicRouteID,
			RouteName:       routeName,
			ParamsTypeName:  routeName + "Params",
			Params:          params,
		}
	}

	for routeID, notFound := range notFounds {
		if _, ok := contractsByRoute[routeID]; ok {
			continue
		}
		params, err := routeParamsFromSegments(notFound.PublicRouteID, notFound.PublicSegments)
		if err != nil {
			return nil, err
		}
		routeName := routeNameFromSegments(notFound.Segments)
		contractsByRoute[routeID] = routeContractDef{
			RouteID:         routeID,
			InternalRouteID: notFound.InternalRouteID,
			PublicRouteID:   notFound.PublicRouteID,
			RouteName:       routeName,
			ParamsTypeName:  routeName + "Params",
			Params:          params,
		}
	}

	if _, ok := contractsByRoute[""]; !ok {
		contractsByRoute[""] = routeContractDef{
			RouteID:         "",
			InternalRouteID: "",
			PublicRouteID:   "",
			RouteName:       "Root",
			ParamsTypeName:  "RootParams",
		}
	}

	routeIDs := make([]string, 0, len(contractsByRoute))
	for routeID := range contractsByRoute {
		routeIDs = append(routeIDs, routeID)
	}
	sort.Strings(routeIDs)

	contracts := make([]routeContractDef, 0, len(routeIDs))
	for _, routeID := range routeIDs {
		contracts = append(contracts, contractsByRoute[routeID])
	}
	return contracts, nil
}

func contractsByRouteID(contracts []routeContractDef) map[string]routeContractDef {
	byRoute := make(map[string]routeContractDef, len(contracts))
	for _, contract := range contracts {
		byRoute[contract.RouteID] = contract
	}
	return byRoute
}

func mergeTemplateMaps(maps ...map[string]templateDef) map[string]templateDef {
	total := 0
	for _, items := range maps {
		total += len(items)
	}
	out := make(map[string]templateDef, total)
	for _, items := range maps {
		for routeID, tpl := range items {
			out[routeID] = tpl
		}
	}
	return out
}

func sortedTemplateRouteIDs(templates map[string]templateDef) []string {
	routeIDs := make([]string, 0, len(templates))
	for routeID := range templates {
		routeIDs = append(routeIDs, routeID)
	}
	sort.Strings(routeIDs)
	return routeIDs
}

func slotLayoutsFromOwners(slotOwners map[string][]slotDef) map[string]templateDef {
	out := make(map[string]templateDef)
	for _, slots := range slotOwners {
		for _, slot := range slots {
			for routeID, layout := range slot.Layouts {
				out[routeID] = layout
			}
		}
	}
	return out
}

func defaultsFromSlotOwners(slotOwners map[string][]slotDef) map[string]templateDef {
	out := make(map[string]templateDef)
	for _, slots := range slotOwners {
		for _, slot := range slots {
			if slot.Default != nil {
				out[slot.Default.RouteID] = *slot.Default
			}
		}
	}
	return out
}

func clientAssetRouteSpecs(
	root templateDef,
	metas []routeMeta,
	layouts map[string]templateDef,
	slotOwners map[string][]slotDef,
) []clientassets.RouteSpec {
	specs := make([]clientassets.RouteSpec, 0, len(metas))
	for _, meta := range metas {
		templatePaths := []string{}
		seen := map[string]struct{}{}
		cssBundles := []clientassets.CSSBundleSpec{}

		appendClientAssetTemplatePath(&templatePaths, seen, root)
		if strings.TrimSpace(root.SourcePath) != "" {
			cssBundles = append(cssBundles, clientassets.CSSBundleSpec{
				OwnerTemplatePath: root.SourcePath,
				TemplatePaths:     []string{root.SourcePath},
			})
		}

		lastNonRootLayout := -1
		for _, layout := range layoutChain(meta.RouteID, layouts) {
			appendClientAssetTemplatePath(&templatePaths, seen, layout)
			bundlePaths := []string{layout.SourcePath}
			for _, slotTemplate := range clientAssetSlotTemplatesForRoute(slotOwners[layout.RouteID]) {
				appendClientAssetTemplatePath(&templatePaths, seen, slotTemplate)
				bundlePaths = append(bundlePaths, slotTemplate.SourcePath)
			}
			if layout.RouteID != "" {
				lastNonRootLayout = len(cssBundles)
			}
			cssBundles = append(cssBundles, clientassets.CSSBundleSpec{
				OwnerTemplatePath: layout.SourcePath,
				TemplatePaths:     bundlePaths,
			})
		}

		appendClientAssetTemplatePath(&templatePaths, seen, meta.Page)
		if strings.TrimSpace(meta.Page.SourcePath) != "" {
			if lastNonRootLayout >= 0 {
				cssBundles[lastNonRootLayout].TemplatePaths = append(
					cssBundles[lastNonRootLayout].TemplatePaths,
					meta.Page.SourcePath,
				)
			} else {
				cssBundles = append(cssBundles, clientassets.CSSBundleSpec{
					OwnerTemplatePath: meta.Page.SourcePath,
					TemplatePaths:     []string{meta.Page.SourcePath},
				})
			}
		}

		specs = append(specs, clientassets.RouteSpec{
			RouteID:       meta.RouteID,
			TemplatePaths: templatePaths,
			CSSBundles:    cssBundles,
		})
	}
	return specs
}

func clientAssetSlotTemplatesForRoute(slots []slotDef) []templateDef {
	templates := []templateDef{}
	seen := map[string]struct{}{}
	for _, slot := range slots {
		appendClientAssetTemplateDefs(&templates, seen, layoutChain(slot.RootInternal, slot.Layouts)...)
		if slot.Default != nil {
			appendClientAssetTemplateDefs(&templates, seen, *slot.Default)
		}
		for _, page := range slot.Pages {
			appendClientAssetTemplateDefs(&templates, seen, layoutChain(page.RouteID, slot.Layouts)...)
			appendClientAssetTemplateDefs(&templates, seen, page.Page)
		}
	}
	return templates
}

func appendClientAssetTemplateDefs(paths *[]templateDef, seen map[string]struct{}, templates ...templateDef) {
	for _, tpl := range templates {
		if strings.TrimSpace(tpl.SourcePath) == "" {
			continue
		}
		if _, ok := seen[tpl.SourcePath]; ok {
			continue
		}
		seen[tpl.SourcePath] = struct{}{}
		*paths = append(*paths, tpl)
	}
}

func appendClientAssetTemplatePath(paths *[]string, seen map[string]struct{}, tpl templateDef) {
	if strings.TrimSpace(tpl.SourcePath) == "" {
		return
	}
	if _, ok := seen[tpl.SourcePath]; ok {
		return
	}
	seen[tpl.SourcePath] = struct{}{}
	*paths = append(*paths, tpl.SourcePath)
}
