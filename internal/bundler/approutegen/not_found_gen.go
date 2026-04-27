package approutegen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/RevoTale/no-js/framework/metagen"
)

func writeNotFoundPageFunc(
	buffer *bytes.Buffer,
	root templateDef,
	layouts map[string]templateDef,
	notFounds map[string]templateDef,
	slotNames map[string][]string,
	contractsByID map[string]routeContractDef,
	notFoundAssets map[string]metagen.ClientAssets,
) error {
	notFoundKeys := make([]string, 0, len(notFounds))
	dynamicNotFoundKeys := make([]string, 0, len(notFounds))
	for routeID := range notFounds {
		notFoundKeys = append(notFoundKeys, routeID)
		if routeIDHasParams(routeID) {
			dynamicNotFoundKeys = append(dynamicNotFoundKeys, routeID)
		}
	}
	sort.Strings(notFoundKeys)
	sort.Slice(dynamicNotFoundKeys, func(i int, j int) bool {
		left := routeIDSegmentCount(dynamicNotFoundKeys[i])
		right := routeIDSegmentCount(dynamicNotFoundKeys[j])
		if left != right {
			return left > right
		}
		return dynamicNotFoundKeys[i] < dynamicNotFoundKeys[j]
	})

	buffer.WriteString(
		"func renderNotFoundPage(resolvers RouteResolvers, appCtx *view.Context, r *http.Request, " +
			"notFound framework.NotFoundContext) (templ.Component, error) {\n",
	)
	buffer.WriteString("\tpathValue := strings.TrimSpace(notFound.RequestPath)\n")
	buffer.WriteString("\tif pathValue == \"\" {\n")
	buffer.WriteString("\t\tpathValue = \"/\"\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\trouteID := nearestNotFoundRouteID(notFound)\n")
	buffer.WriteString("\tr = withNotFoundRequestInfo(r, notFound)\n")
	buffer.WriteString("\tmeta := metagen.Metadata{\n")
	buffer.WriteString("\t\tRobots: &metagen.Robots{\n")
	buffer.WriteString("\t\t\tIndex: metagen.Bool(false),\n")
	buffer.WriteString("\t\t\tFollow: metagen.Bool(false),\n")
	buffer.WriteString("\t\t},\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tmeta = metagen.MergeManagedStylesheets(requestContext(r), meta)\n")
	buffer.WriteString(
		"\tmeta = metagen.MergeManagedClientAssets(requestContext(r), meta, " +
			"notFoundClientAssets(routeID))\n",
	)
	buffer.WriteString("\tswitch routeID {\n")
	for _, routeID := range notFoundKeys {
		if routeID == "" {
			continue
		}
		notFound := notFounds[routeID]
		contract, ok := contractsByID[routeID]
		if !ok {
			return fmt.Errorf("missing route contract for not-found route %q", routeID)
		}
		writef(buffer, "\tcase %q:\n", routeID)
		writeNotFoundParams(buffer, "\t\t", contract)
		writef(
			buffer,
			"\t\tview, err := resolvers.%s(requestContext(r), appCtx, r, notFound, params)\n",
			resolveNotFoundMethod(notFound),
		)
		buffer.WriteString("\t\tif err != nil {\n")
		buffer.WriteString("\t\t\treturn nil, err\n")
		buffer.WriteString("\t\t}\n")
		writef(buffer, "\t\tcomponent := %s.NotFound(view, pathValue)\n", notFound.ModuleName)
		chain := layoutChain(routeID, layouts)
		layoutModelVars, err := writeResolveLayoutModels(
			buffer,
			"\t\t",
			chain,
			contractsByID,
			"requestContext(r)",
			"params",
			contract.Params,
			"appCtx",
		)
		if err != nil {
			return err
		}
		for idx := len(chain) - 1; idx >= 0; idx-- {
			layout := chain[idx]
			expr := layoutInvocationExpr(
				layout,
				slotNames[layout.RouteID],
				"meta",
				layoutModelVars[layout.RouteID],
				"component",
				nil,
			)
			writef(buffer, "\t\tcomponent = %s\n", expr)
		}
		writef(buffer, "\t\treturn %s.RootLayout(meta, notFound.Locale, component), nil\n", root.ModuleName)
	}

	rootNotFound := notFounds[""]
	rootContract, ok := contractsByID[""]
	if !ok {
		return fmt.Errorf("missing route contract for root not-found route")
	}
	buffer.WriteString("\tdefault:\n")
	writeNotFoundParams(buffer, "\t\t", rootContract)
	writef(
		buffer,
		"\t\tview, err := resolvers.%s(requestContext(r), appCtx, r, notFound, params)\n",
		resolveNotFoundMethod(rootNotFound),
	)
	buffer.WriteString("\t\tif err != nil {\n")
	buffer.WriteString("\t\t\treturn nil, err\n")
	buffer.WriteString("\t\t}\n")
	writef(buffer, "\t\tcomponent := %s.NotFound(view, pathValue)\n", rootNotFound.ModuleName)
	rootChain := layoutChain("", layouts)
	layoutModelVars, err := writeResolveLayoutModels(
		buffer,
		"\t\t",
		rootChain,
		contractsByID,
		"requestContext(r)",
		"params",
		rootContract.Params,
		"appCtx",
	)
	if err != nil {
		return err
	}
	for idx := len(rootChain) - 1; idx >= 0; idx-- {
		layout := rootChain[idx]
		expr := layoutInvocationExpr(
			layout,
			slotNames[layout.RouteID],
			"meta",
			layoutModelVars[layout.RouteID],
			"component",
			nil,
		)
		writef(buffer, "\t\tcomponent = %s\n", expr)
	}
	writef(buffer, "\t\treturn %s.RootLayout(meta, notFound.Locale, component), nil\n", root.ModuleName)
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n\n")

	writeClientAssetsSwitchFunc(buffer, "notFoundClientAssets", notFoundAssets)

	buffer.WriteString("func nearestNotFoundRouteID(notFound framework.NotFoundContext) string {\n")
	buffer.WriteString("\tfor _, candidate := range routeAncestry(notFound.MatchedRouteID) {\n")
	buffer.WriteString("\t\tif routeID, ok := resolveNotFoundCandidateRouteID(candidate); ok {\n")
	buffer.WriteString("\t\t\treturn routeID\n")
	buffer.WriteString("\t\t}\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tfor _, candidate := range candidateRouteIDsFromPattern(notFound.MatchedRoutePattern) {\n")
	buffer.WriteString("\t\tif routeID, ok := resolveNotFoundCandidateRouteID(candidate); ok {\n")
	buffer.WriteString("\t\t\treturn routeID\n")
	buffer.WriteString("\t\t}\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tfor _, candidate := range candidateRouteIDsFromPath(notFound.RequestPath) {\n")
	buffer.WriteString("\t\tif routeID, ok := resolveNotFoundCandidateRouteID(candidate); ok {\n")
	buffer.WriteString("\t\t\treturn routeID\n")
	buffer.WriteString("\t\t}\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\treturn \"\"\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func resolveNotFoundCandidateRouteID(candidate string) (string, bool) {\n")
	buffer.WriteString("\tif hasNotFoundTemplate(candidate) {\n")
	buffer.WriteString("\t\treturn candidate, true\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tif candidate == \"\" {\n")
	buffer.WriteString("\t\treturn \"\", false\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\trouteID, ok := matchDynamicNotFoundTemplate(candidate)\n")
	buffer.WriteString("\tif !ok {\n")
	buffer.WriteString("\t\treturn \"\", false\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\treturn routeID, true\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func matchDynamicNotFoundTemplate(candidate string) (string, bool) {\n")
	buffer.WriteString("\tif candidate == \"\" {\n")
	buffer.WriteString("\t\treturn \"\", false\n")
	buffer.WriteString("\t}\n")
	if len(dynamicNotFoundKeys) > 0 {
		buffer.WriteString("\trequestPath := \"/\" + candidate\n")
	}
	for _, routeID := range dynamicNotFoundKeys {
		writef(
			buffer,
			"\tif _, ok := router.MatchPathPattern(%q, requestPath); ok {\n",
			routePattern(notFounds[routeID].PublicRouteID),
		)
		writef(buffer, "\t\treturn %q, true\n", routeID)
		buffer.WriteString("\t}\n")
	}
	buffer.WriteString("\treturn \"\", false\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func hasNotFoundTemplate(routeID string) bool {\n")
	buffer.WriteString("\tswitch routeID {\n")
	if len(notFoundKeys) > 0 {
		buffer.WriteString("\tcase ")
		for idx, routeID := range notFoundKeys {
			if idx > 0 {
				buffer.WriteString(", ")
			}
			writef(buffer, "%q", routeID)
		}
		buffer.WriteString(":\n")
	}
	buffer.WriteString("\t\treturn true\n")
	buffer.WriteString("\tdefault:\n")
	buffer.WriteString("\t\treturn false\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func candidateRouteIDsFromPattern(pattern string) []string {\n")
	buffer.WriteString("\trouteID := normalizePatternRouteID(pattern)\n")
	buffer.WriteString("\treturn routeAncestry(routeID)\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func candidateRouteIDsFromPath(requestPath string) []string {\n")
	buffer.WriteString("\trouteID := normalizeRequestRouteID(requestPath)\n")
	buffer.WriteString("\treturn routeAncestry(routeID)\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func normalizePatternRouteID(pattern string) string {\n")
	buffer.WriteString("\trouteID := strings.TrimSpace(pattern)\n")
	buffer.WriteString("\trouteID = strings.Trim(routeID, \"/\")\n")
	buffer.WriteString("\treturn routeID\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func normalizeRequestRouteID(requestPath string) string {\n")
	buffer.WriteString("\trouteID := strings.TrimSpace(requestPath)\n")
	buffer.WriteString("\trouteID = strings.Trim(routeID, \"/\")\n")
	buffer.WriteString("\treturn routeID\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func routeAncestry(routeID string) []string {\n")
	buffer.WriteString("\trouteID = strings.TrimSpace(routeID)\n")
	buffer.WriteString("\trouteID = strings.Trim(routeID, \"/\")\n")
	buffer.WriteString("\tif routeID == \"\" {\n")
	buffer.WriteString("\t\treturn []string{\"\"}\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tparts := strings.Split(routeID, \"/\")\n")
	buffer.WriteString("\tout := make([]string, 0, len(parts)+1)\n")
	buffer.WriteString("\tfor idx := len(parts); idx >= 1; idx-- {\n")
	buffer.WriteString("\t\tout = append(out, strings.Join(parts[:idx], \"/\"))\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tout = append(out, \"\")\n")
	buffer.WriteString("\treturn out\n")
	buffer.WriteString("}\n\n")
	return nil
}

func routeIDHasParams(routeID string) bool {
	for _, segment := range segmentsFromRouteID(routeID) {
		if segment.IsParam() {
			return true
		}
	}
	return false
}

func routeIDSegmentCount(routeID string) int {
	routeID = strings.TrimSpace(routeID)
	routeID = strings.Trim(routeID, "/")
	if routeID == "" {
		return 0
	}
	return len(strings.Split(routeID, "/"))
}

type paramAssignment struct {
	TargetField string
	SourceField string
}

func mapPageParamsByName(params []routeParamDef) map[string]routeParamDef {
	byName := make(map[string]routeParamDef, len(params))
	for _, param := range params {
		byName[param.Name] = param
	}
	return byName
}

func layoutParamAssignments(
	pageParamsByName map[string]routeParamDef,
	layoutParams []routeParamDef,
) ([]paramAssignment, error) {
	assignments := make([]paramAssignment, 0, len(layoutParams))
	for _, layoutParam := range layoutParams {
		pageParam, ok := pageParamsByName[layoutParam.Name]
		if !ok {
			return nil, fmt.Errorf("missing param %q on page route", layoutParam.Name)
		}
		assignments = append(assignments, paramAssignment{
			TargetField: layoutParam.FieldName,
			SourceField: pageParam.FieldName,
		})
	}
	return assignments, nil
}

func writeParamsAssignment(
	buffer *bytes.Buffer,
	indent string,
	targetVar string,
	targetTypeName string,
	sourceVar string,
	sourceParams []routeParamDef,
	targetParams []routeParamDef,
) error {
	writef(buffer, "%s%s := route_resolvers.%s{}\n", indent, targetVar, targetTypeName)
	if len(targetParams) == 0 {
		return nil
	}
	assignments, err := layoutParamAssignments(mapPageParamsByName(sourceParams), targetParams)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		writef(buffer, "%s%s.%s = %s.%s\n", indent, targetVar, assignment.TargetField, sourceVar, assignment.SourceField)
	}
	return nil
}

func writeNotFoundParams(buffer *bytes.Buffer, indent string, contract routeContractDef) {
	writef(buffer, "%sparams, _ := %s(notFoundStrippedPath(notFound))\n", indent, parseParamsFuncNameForContract(contract))
}

func layoutModelVarName(layout templateDef) string {
	return strings.ToLower(routeNameFromSegments(layout.Segments)) + "LayoutView"
}

func layoutParamsVarName(layout templateDef) string {
	return strings.ToLower(routeNameFromSegments(layout.Segments)) + "LayoutParams"
}

func defaultModelVarName(fallback templateDef) string {
	return strings.ToLower(routeNameFromSegments(fallback.Segments)) + "DefaultView"
}

func defaultParamsVarName(fallback templateDef) string {
	return strings.ToLower(routeNameFromSegments(fallback.Segments)) + "DefaultParams"
}

func writeResolveLayoutModels(
	buffer *bytes.Buffer,
	indent string,
	chain []templateDef,
	contractsByID map[string]routeContractDef,
	ctxExpr string,
	sourceParamsVar string,
	sourceParams []routeParamDef,
	appCtxExpr string,
) (map[string]string, error) {
	modelVars := make(map[string]string, len(chain))
	for _, layout := range chain {
		contract, ok := contractsByID[layout.RouteID]
		if !ok {
			return nil, fmt.Errorf("missing route contract for layout route %q", layout.RouteID)
		}
		paramsVar := layoutParamsVarName(layout)
		if err := writeParamsAssignment(
			buffer,
			indent,
			paramsVar,
			contract.ParamsTypeName,
			sourceParamsVar,
			sourceParams,
			contract.Params,
		); err != nil {
			return nil, fmt.Errorf("layout %q params: %w", layout.RouteID, err)
		}
		modelVar := layoutModelVarName(layout)
		writef(
			buffer,
			"%s%s, err := resolvers.%s(%s, %s, r, %s)\n",
			indent,
			modelVar,
			resolveLayoutMethod(layout),
			ctxExpr,
			appCtxExpr,
			paramsVar,
		)
		writef(buffer, "%sif err != nil {\n", indent)
		writef(buffer, "%s\treturn nil, err\n", indent)
		writef(buffer, "%s}\n", indent)
		modelVars[layout.RouteID] = modelVar
	}
	return modelVars, nil
}

func writeClientAssetsSwitchFunc(
	buffer *bytes.Buffer,
	funcName string,
	assetsByRoute map[string]metagen.ClientAssets,
) {
	buffer.WriteString("func " + funcName + "(routeID string) metagen.ClientAssets {\n")
	keys := make([]string, 0, len(assetsByRoute))
	for routeID, assets := range assetsByRoute {
		if clientAssetsEmpty(assets) {
			continue
		}
		keys = append(keys, routeID)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		buffer.WriteString("\tswitch routeID {\n")
		for _, routeID := range keys {
			writef(buffer, "\tcase %q:\n", routeID)
			writeClientAssetsReturn(buffer, "\t\t", assetsByRoute[routeID])
		}
		buffer.WriteString("\t}\n")
	}
	buffer.WriteString("\treturn metagen.ClientAssets{}\n")
	buffer.WriteString("}\n\n")
}

func writeClientAssetsLiteral(buffer *bytes.Buffer, indent string, assets metagen.ClientAssets) {
	if clientAssetsEmpty(assets) {
		return
	}
	buffer.WriteString(indent + "ClientAssets: metagen.ClientAssets{\n")
	writeClientAssetsFields(buffer, indent+"\t", assets)
	buffer.WriteString(indent + "},\n")
}

func writeClientAssetsReturn(buffer *bytes.Buffer, indent string, assets metagen.ClientAssets) {
	buffer.WriteString(indent + "return metagen.ClientAssets{\n")
	writeClientAssetsFields(buffer, indent+"\t", assets)
	buffer.WriteString(indent + "}\n")
}

func writeClientAssetsFields(buffer *bytes.Buffer, indent string, assets metagen.ClientAssets) {
	if len(assets.Stylesheets) > 0 {
		buffer.WriteString(indent + "Stylesheets: []string{\n")
		for _, stylesheet := range assets.Stylesheets {
			writef(buffer, indent+"\t%q,\n", stylesheet)
		}
		buffer.WriteString(indent + "},\n")
	}
	if len(assets.ModuleScripts) > 0 {
		buffer.WriteString(indent + "ModuleScripts: []string{\n")
		for _, script := range assets.ModuleScripts {
			writef(buffer, indent+"\t%q,\n", script)
		}
		buffer.WriteString(indent + "},\n")
	}
}

func clientAssetsEmpty(assets metagen.ClientAssets) bool {
	return len(assets.Stylesheets) == 0 && len(assets.ModuleScripts) == 0
}
