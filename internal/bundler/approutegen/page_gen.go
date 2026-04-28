package approutegen

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/RevoTale/no-js/framework/metagen"
	frameworkrouter "github.com/RevoTale/no-js/framework/router"
)

func writePageModule(
	buffer *bytes.Buffer,
	meta routeMeta,
	root templateDef,
	layouts map[string]templateDef,
	contractsByID map[string]routeContractDef,
	routeAssets map[string]metagen.ClientAssets,
) error {
	if _, ok := contractsByID[meta.RouteID]; !ok {
		return fmt.Errorf("missing route contract for route %q", meta.RouteID)
	}

	writef(
		buffer,
		"\t\t\tPage: framework.PageModule[*view.Context, %s, %s]{\n",
		meta.ParamsTypeName,
		meta.PageViewType,
	)
	writef(buffer, "\t\t\t\tRouteID:      %q,\n", meta.RouteID)
	writef(buffer, "\t\t\t\tPattern:     %q,\n", routePattern(meta.PublicRouteID))
	writef(buffer, "\t\t\t\tParseParams: %s,\n", parseParamsFuncName(meta))
	writef(
		buffer,
		"\t\t\t\tMetaGenContext: func(meta framework.MetaContext[*view.Context], "+
			"params %s) (metagen.Metadata, error) {\n",
		meta.ParamsTypeName,
	)
	writef(buffer, "\t\t\t\t\treturn resolvers.%s(meta, params)\n", metaGenPageMethod(meta))
	buffer.WriteString("\t\t\t\t},\n")
	chain := layoutChain(meta.RouteID, layouts)
	writef(buffer, "\t\t\t\tMetaGenName: %q,\n", resolverMethodQualified(metaGenPageMethod(meta)))
	metaChainMethodNames := make([]string, 0, len(chain)+2)
	metaChainMethodNames = append(
		metaChainMethodNames,
		resolverMethodQualified(metaGenLayoutMethod(templateDef{RouteID: ""})),
	)
	for _, layout := range chain {
		if layout.RouteID == "" {
			continue
		}
		metaChainMethodNames = append(metaChainMethodNames, resolverMethodQualified(metaGenLayoutMethod(layout)))
	}
	metaChainMethodNames = append(metaChainMethodNames, resolverMethodQualified(metaGenPageMethod(meta)))
	buffer.WriteString("\t\t\t\tMetaGenChainNames: []string{\n")
	for _, methodName := range metaChainMethodNames {
		writef(buffer, "\t\t\t\t\t%q,\n", methodName)
	}
	buffer.WriteString("\t\t\t\t},\n")
	writef(
		buffer,
		"\t\t\t\tMetaGenContextChain: []framework.PageMetaGenContext[*view.Context, %s]{\n",
		meta.ParamsTypeName,
	)
	writef(
		buffer,
		"\t\t\t\t\tfunc(meta framework.MetaContext[*view.Context], _ %s) "+
			"(metagen.Metadata, error) {\n",
		meta.ParamsTypeName,
	)
	buffer.WriteString("\t\t\t\t\t\treturn resolvers.MetaGenRootLayout(meta)\n")
	buffer.WriteString("\t\t\t\t\t},\n")
	for _, layout := range chain {
		if layout.RouteID == "" {
			continue
		}
		contract, ok := contractsByID[layout.RouteID]
		if !ok {
			return fmt.Errorf("missing route contract for layout route %q", layout.RouteID)
		}
		pageParams := mapPageParamsByName(meta.Params)
		assignments, err := layoutParamAssignments(pageParams, contract.Params)
		if err != nil {
			return fmt.Errorf("route %q layout %q metadata params: %w", meta.RouteID, layout.RouteID, err)
		}
		writef(
			buffer,
			"\t\t\t\t\tfunc(meta framework.MetaContext[*view.Context], "+
				"params %s) (metagen.Metadata, error) {\n",
			meta.ParamsTypeName,
		)
		if len(contract.Params) > 0 {
			writef(buffer, "\t\t\t\t\t\tlayoutParams := route_resolvers.%s{}\n", contract.ParamsTypeName)
			for _, assignment := range assignments {
				writef(buffer, "\t\t\t\t\t\tlayoutParams.%s = params.%s\n", assignment.TargetField, assignment.SourceField)
			}
			writef(
				buffer,
				"\t\t\t\t\t\treturn resolvers.%s(meta, layoutParams)\n",
				metaGenLayoutMethod(layout),
			)
		} else {
			writef(
				buffer,
				"\t\t\t\t\t\treturn resolvers.%s(meta, route_resolvers.%s{})\n",
				metaGenLayoutMethod(layout),
				contract.ParamsTypeName,
			)
		}
		buffer.WriteString("\t\t\t\t\t},\n")
	}
	writef(
		buffer,
		"\t\t\t\t\tfunc(meta framework.MetaContext[*view.Context], "+
			"params %s) (metagen.Metadata, error) {\n",
		meta.ParamsTypeName,
	)
	writef(buffer, "\t\t\t\t\t\treturn resolvers.%s(meta, params)\n", metaGenPageMethod(meta))
	buffer.WriteString("\t\t\t\t\t},\n")
	buffer.WriteString("\t\t\t\t},\n")
	writef(
		buffer,
		"\t\t\t\tLoad: func(ctx context.Context, appCtx *view.Context, r *http.Request, "+
			"params %s) (%s, error) {\n",
		meta.ParamsTypeName,
		meta.PageViewType,
	)
	writef(buffer, "\t\t\t\t\treturn resolvers.%s(ctx, appCtx, r, params)\n", resolvePageMethod(meta))
	buffer.WriteString("\t\t\t\t},\n")
	writef(buffer, "\t\t\t\tLoadName: %q,\n", resolverMethodQualified(resolvePageMethod(meta)))
	writef(
		buffer,
		"\t\t\t\tCompose: func(ctx context.Context, runtime framework.RuntimeContext[*view.Context], "+
			"r *http.Request, meta metagen.Metadata, view %s, params %s, partial bool) (templ.Component, error) {\n",
		meta.PageViewType,
		meta.ParamsTypeName,
	)
	writef(
		buffer,
		"\t\t\t\t\treturn %s(ctx, runtime, r, meta, view, params, partial, resolvers)\n",
		composeFuncName(meta),
	)
	buffer.WriteString("\t\t\t\t},\n")
	writef(buffer, "\t\t\t\tRender: %s.Page,\n", meta.Page.ModuleName)
	writef(buffer, "\t\t\t\tRootLayout: %s.RootLayout,\n", root.ModuleName)
	writeClientAssetsLiteral(buffer, "\t\t\t\t", routeAssets[meta.RouteID])
	buffer.WriteString("\t\t\t},\n")
	return nil
}

func parseParamsFuncName(meta routeMeta) string {
	return "parse" + meta.RouteName + "Params"
}

func parseParamsFuncNameForContract(contract routeContractDef) string {
	return "parse" + contract.RouteName + "Params"
}

func writeParseParamsFunc(buffer *bytes.Buffer, contract routeContractDef) {
	funcName := parseParamsFuncNameForContract(contract)
	pattern := routePattern(contract.PublicRouteID)

	writef(buffer, "func %s(requestPath string) (route_resolvers.%s, bool) {\n", funcName, contract.ParamsTypeName)
	if len(contract.Params) == 0 {
		writef(buffer, "\t_, ok := router.MatchPathPattern(%q, requestPath)\n", pattern)
		buffer.WriteString("\tif !ok {\n")
		writef(buffer, "\t\treturn route_resolvers.%s{}, false\n", contract.ParamsTypeName)
		buffer.WriteString("\t}\n")
		writef(buffer, "\treturn route_resolvers.%s{}, true\n", contract.ParamsTypeName)
		buffer.WriteString("}\n\n")
		return
	}

	writef(buffer, "\tparams, ok := router.MatchPathPattern(%q, requestPath)\n", pattern)
	buffer.WriteString("\tif !ok {\n")
	writef(buffer, "\t\treturn route_resolvers.%s{}, false\n", contract.ParamsTypeName)
	buffer.WriteString("\t}\n")
	writef(buffer, "\tout := route_resolvers.%s{}\n", contract.ParamsTypeName)
	for _, param := range contract.Params {
		switch param.Type {
		case "[]string":
			writef(buffer, "\tif %sValue, exists := params[%q]; exists {\n", param.FieldName, param.Name)
			writef(buffer, "\t\tout.%s = slices.Clone(%sValue)\n", param.FieldName, param.FieldName)
			buffer.WriteString("\t}\n")
		default:
			writef(buffer, "\t%sValue, exists := params[%q]\n", param.FieldName, param.Name)
			buffer.WriteString("\tif !exists || len(" + param.FieldName + "Value) == 0 {\n")
			writef(buffer, "\t\treturn route_resolvers.%s{}, false\n", contract.ParamsTypeName)
			buffer.WriteString("\t}\n")
			writef(buffer, "\tout.%s = strings.TrimSpace(%sValue[0])\n", param.FieldName, param.FieldName)
		}
	}
	buffer.WriteString("\treturn out, true\n")
	buffer.WriteString("}\n\n")
}

func routePattern(routeID string) string {
	if routeID == "" {
		return "/"
	}
	return "/" + routeID
}

func segmentsFromRouteID(routeID string) []routeSegment {
	trimmed := strings.Trim(strings.TrimSpace(routeID), "/")
	if trimmed == "" {
		return nil
	}

	rawSegments, err := frameworkrouter.ParseDirectorySegments(trimmed)
	if err != nil {
		return nil
	}
	segments := make([]routeSegment, 0, len(rawSegments))
	for _, raw := range rawSegments {
		segments = append(segments, newRouteSegment(raw))
	}
	return segments
}

func discoveryImportAlias(routeID string) string {
	if strings.TrimSpace(routeID) == "" {
		return "route_conventions_root"
	}
	return "route_conventions_" + safeIdentifier(routeID)
}

func conventionEndpointPattern(routeID string, leaf string) string {
	leaf = strings.Trim(strings.TrimSpace(leaf), "/")
	if leaf == "" {
		return routePattern(routeID)
	}
	if strings.TrimSpace(routeID) == "" {
		return "/" + leaf
	}
	return "/" + path.Join(strings.Trim(routeID, "/"), leaf)
}

func resolvePageMethod(meta routeMeta) string {
	return "Resolve" + meta.RouteName + "Page"
}

func resolveLayoutMethod(layout templateDef) string {
	return "Resolve" + routeNameFromSegments(layout.Segments) + "Layout"
}

func resolveDefaultMethod(fallback templateDef) string {
	return "Resolve" + routeNameFromSegments(fallback.Segments) + "Default"
}

func resolveNotFoundMethod(notFound templateDef) string {
	return "Resolve" + routeNameFromSegments(notFound.Segments) + "NotFound"
}

func metaGenPageMethod(meta routeMeta) string {
	return "MetaGen" + meta.RouteName + "Page"
}

func metaGenNotFoundMethod(notFound templateDef) string {
	return "MetaGen" + routeNameFromSegments(notFound.Segments) + "NotFound"
}

func metaGenLayoutMethod(layout templateDef) string {
	if layout.RouteID == "" {
		return "MetaGenRootLayout"
	}
	return "MetaGen" + routeNameFromSegments(layout.Segments) + "Layout"
}

func resolverMethodQualified(method string) string {
	return "route_resolvers.Resolver." + strings.TrimSpace(method)
}

func layoutChain(routeID string, layouts map[string]templateDef) []templateDef {
	segments := []string{}
	if routeID != "" {
		segments = strings.Split(routeID, "/")
	}

	candidates := make([]string, 0, len(segments)+1)
	candidates = append(candidates, "")
	for idx := 1; idx <= len(segments); idx++ {
		candidates = append(candidates, strings.Join(segments[:idx], "/"))
	}

	chain := make([]templateDef, 0, len(candidates))
	for _, candidate := range candidates {
		layout, ok := layouts[candidate]
		if !ok {
			continue
		}
		chain = append(chain, layout)
	}

	return chain
}

func composeFuncName(meta routeMeta) string {
	return "compose" + meta.RouteName + "Page"
}

func slotResolveFuncName(ownerRouteID string, slotName string) string {
	return "resolve" + routeNameFromSegments(segmentsFromRouteID(ownerRouteID)) + pascalToken(slotName) + "Slot"
}

func slotComponentVarName(ownerRouteID string, slotName string) string {
	return strings.ToLower(routeNameFromSegments(segmentsFromRouteID(ownerRouteID))) + pascalToken(slotName) + "Slot"
}

func layoutInvocationExpr(
	layout templateDef,
	slotNames []string,
	metaExpr string,
	viewExpr string,
	childExpr string,
	slotExprs map[string]string,
) string {
	args := make([]string, 0, len(slotNames)+3)
	if layout.RouteID == "" {
		args = append(args, metaExpr)
	}
	args = append(args, viewExpr, childExpr)
	for _, slotName := range slotNames {
		value := "nil"
		if slotExprs != nil {
			if expr, ok := slotExprs[slotName]; ok {
				value = expr
			}
		}
		args = append(args, value)
	}
	return layout.ModuleName + ".Layout(" + strings.Join(args, ", ") + ")"
}
