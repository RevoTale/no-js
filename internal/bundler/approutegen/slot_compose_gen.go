package approutegen

import (
	"bytes"
	"fmt"
	"slices"
	"sort"

	frameworkrouter "github.com/RevoTale/no-js/framework/router"
)

func writeSlotResolveFuncs(
	buffer *bytes.Buffer,
	slotOwners map[string][]slotDef,
	slotNames map[string][]string,
	contractsByID map[string]routeContractDef,
) error {
	ownerRouteIDs := make([]string, 0, len(slotOwners))
	for ownerRouteID := range slotOwners {
		ownerRouteIDs = append(ownerRouteIDs, ownerRouteID)
	}
	sort.Strings(ownerRouteIDs)

	for _, ownerRouteID := range ownerRouteIDs {
		slots := slices.Clone(slotOwners[ownerRouteID])
		sort.Slice(slots, func(i int, j int) bool {
			return slots[i].Name < slots[j].Name
		})
		for _, slot := range slots {
			if err := writeSlotResolveFunc(buffer, ownerRouteID, slot, slotNames, contractsByID); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSlotResolveFunc(
	buffer *bytes.Buffer,
	ownerRouteID string,
	slot slotDef,
	slotNames map[string][]string,
	contractsByID map[string]routeContractDef,
) error {
	funcName := slotResolveFuncName(ownerRouteID, slot.Name)
	ownerContract, ok := contractsByID[ownerRouteID]
	if !ok {
		return fmt.Errorf("missing route contract for slot owner route %q", ownerRouteID)
	}
	writef(
		buffer,
		"func %s(ctx context.Context, runtime framework.RuntimeContext[*view.Context], "+
			"r *http.Request, params route_resolvers.%s, resolvers RouteResolvers) "+
			"(templ.Component, error) {\n",
		funcName,
		ownerContract.ParamsTypeName,
	)

	pages := slices.Clone(slot.Pages)
	sort.Slice(pages, func(i int, j int) bool {
		left := compareRouteSegmentSpecificity(pages[i].PublicSegments, pages[j].PublicSegments)
		if left != 0 {
			return left > 0
		}
		return pages[i].RouteID < pages[j].RouteID
	})
	for _, page := range pages {
		writef(buffer, "\tif params, ok := %s(r.URL.Path); ok {\n", parseParamsFuncName(page))
		writef(
			buffer,
			"\t\tslotView, err := resolvers.%s(ctx, runtime.AppContext(), r, params)\n",
			resolvePageMethod(page),
		)
		buffer.WriteString("\t\tif err != nil {\n")
		buffer.WriteString("\t\t\treturn nil, err\n")
		buffer.WriteString("\t\t}\n")
		writef(buffer, "\t\tcomponent := %s.Page(slotView)\n", page.Page.ModuleName)
		chain := layoutChain(page.RouteID, slot.Layouts)
		layoutModelVars, err := writeResolveLayoutModels(
			buffer,
			"\t\t",
			chain,
			contractsByID,
			"ctx",
			"params",
			page.Params,
			"runtime.AppContext()",
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
		buffer.WriteString("\t\treturn component, nil\n")
		buffer.WriteString("\t}\n")
	}

	if slot.Default != nil {
		defaultContract, ok := contractsByID[slot.Default.RouteID]
		if !ok {
			return fmt.Errorf("missing route contract for slot default route %q", slot.Default.RouteID)
		}
		defaultParamsVar := defaultParamsVarName(*slot.Default)
		if err := writeParamsAssignment(
			buffer,
			"\t",
			defaultParamsVar,
			defaultContract.ParamsTypeName,
			"params",
			ownerContract.Params,
			defaultContract.Params,
		); err != nil {
			return fmt.Errorf("default %q params: %w", slot.Default.RouteID, err)
		}
		defaultModelVar := defaultModelVarName(*slot.Default)
		writef(
			buffer,
			"\t%s, err := resolvers.%s(ctx, runtime.AppContext(), r, %s)\n",
			defaultModelVar,
			resolveDefaultMethod(*slot.Default),
			defaultParamsVar,
		)
		buffer.WriteString("\tif err != nil {\n")
		buffer.WriteString("\t\treturn nil, err\n")
		buffer.WriteString("\t}\n")
		writef(buffer, "\tcomponent := %s.Default(%s)\n", slot.Default.ModuleName, defaultModelVar)
		chain := layoutChain(slot.RootInternal, slot.Layouts)
		layoutModelVars, err := writeResolveLayoutModels(
			buffer,
			"\t",
			chain,
			contractsByID,
			"ctx",
			defaultParamsVar,
			defaultContract.Params,
			"runtime.AppContext()",
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
			writef(buffer, "\tcomponent = %s\n", expr)
		}
		buffer.WriteString("\treturn component, nil\n")
	} else {
		buffer.WriteString("\treturn nil, nil\n")
	}
	buffer.WriteString("}\n\n")
	return nil
}

func writeComposeFunc(
	buffer *bytes.Buffer,
	meta routeMeta,
	slotOwners map[string][]slotDef,
	slotNames map[string][]string,
	layouts map[string]templateDef,
	contractsByID map[string]routeContractDef,
) error {
	chain := layoutChain(meta.RouteID, layouts)
	writef(
		buffer,
		"func %s(ctx context.Context, runtime framework.RuntimeContext[*view.Context], "+
			"r *http.Request, meta metagen.Metadata, view %s, params %s, partial bool, "+
			"resolvers RouteResolvers) (templ.Component, error) {\n",
		composeFuncName(meta),
		meta.PageViewType,
		meta.ParamsTypeName,
	)
	writef(buffer, "\t_ = params\n")
	writef(buffer, "\tcomponent := %s.Page(view)\n", meta.Page.ModuleName)
	buffer.WriteString("\tif partial {\n")
	buffer.WriteString("\t\treturn component, nil\n")
	buffer.WriteString("\t}\n")
	layoutModelVars, err := writeResolveLayoutModels(
		buffer,
		"\t",
		chain,
		contractsByID,
		"ctx",
		"params",
		meta.Params,
		"runtime.AppContext()",
	)
	if err != nil {
		return err
	}

	hasSlots := false
	for _, layout := range chain {
		if len(slotOwners[layout.RouteID]) > 0 {
			hasSlots = true
			break
		}
	}
	if hasSlots {
		buffer.WriteString("\tslotCtx, cancel := context.WithCancel(ctx)\n")
		buffer.WriteString("\tdefer cancel()\n")
		buffer.WriteString("\tvar slotWG sync.WaitGroup\n")
		buffer.WriteString("\tvar slotErr error\n")
		buffer.WriteString("\tvar slotErrOnce sync.Once\n")
		buffer.WriteString("\tsetSlotErr := func(err error) {\n")
		buffer.WriteString("\t\tif err == nil {\n")
		buffer.WriteString("\t\t\treturn\n")
		buffer.WriteString("\t\t}\n")
		buffer.WriteString("\t\tslotErrOnce.Do(func() {\n")
		buffer.WriteString("\t\t\tslotErr = err\n")
		buffer.WriteString("\t\t\tcancel()\n")
		buffer.WriteString("\t\t})\n")
		buffer.WriteString("\t}\n")
	}

	for _, layout := range chain {
		slots := slotOwners[layout.RouteID]
		if len(slots) == 0 {
			continue
		}
		for _, slot := range slots {
			varName := slotComponentVarName(layout.RouteID, slot.Name)
			writef(buffer, "\tvar %s templ.Component\n", varName)
			if hasSlots {
				buffer.WriteString("\tslotWG.Add(1)\n")
				buffer.WriteString("\tgo func() {\n")
				buffer.WriteString("\t\tdefer slotWG.Done()\n")
				writef(
					buffer,
					"\t\tcomponent, err := %s(slotCtx, runtime, r, %s, resolvers)\n",
					slotResolveFuncName(layout.RouteID, slot.Name),
					layoutParamsVarName(layout),
				)
				buffer.WriteString("\t\tif err != nil {\n")
				buffer.WriteString("\t\t\tsetSlotErr(err)\n")
				buffer.WriteString("\t\t\treturn\n")
				buffer.WriteString("\t\t}\n")
				writef(buffer, "\t\t%s = component\n", varName)
				buffer.WriteString("\t}()\n")
			} else {
				writef(
					buffer,
					"\t%s, err := %s(ctx, runtime, r, %s, resolvers)\n",
					varName,
					slotResolveFuncName(layout.RouteID, slot.Name),
					layoutParamsVarName(layout),
				)
				buffer.WriteString("\tif err != nil {\n")
				buffer.WriteString("\t\treturn nil, err\n")
				buffer.WriteString("\t}\n")
			}
		}
	}
	if hasSlots {
		buffer.WriteString("\tslotWG.Wait()\n")
		buffer.WriteString("\tif slotErr != nil {\n")
		buffer.WriteString("\t\treturn nil, slotErr\n")
		buffer.WriteString("\t}\n")
	}

	for idx := len(chain) - 1; idx >= 0; idx-- {
		layout := chain[idx]
		slotExprs := make(map[string]string)
		for _, slot := range slotOwners[layout.RouteID] {
			slotExprs[slot.Name] = slotComponentVarName(layout.RouteID, slot.Name)
		}
		expr := layoutInvocationExpr(
			layout,
			slotNames[layout.RouteID],
			"meta",
			layoutModelVars[layout.RouteID],
			"component",
			slotExprs,
		)
		writef(buffer, "\tcomponent = %s\n", expr)
	}
	buffer.WriteString("\treturn component, nil\n")
	buffer.WriteString("}\n\n")
	return nil
}

func compareRouteSegmentSpecificity(left []routeSegment, right []routeSegment) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for idx := 0; idx < limit; idx++ {
		leftWeight := routeSegmentSpecificity(left[idx])
		rightWeight := routeSegmentSpecificity(right[idx])
		if leftWeight == rightWeight {
			continue
		}
		if leftWeight > rightWeight {
			return 1
		}
		return -1
	}

	switch {
	case len(left) == len(right):
		return 0
	case len(left) < len(right):
		if remainingOnlyOptionalCatchAllRouteSegments(right[limit:]) {
			return 1
		}
		return -1
	default:
		if remainingOnlyOptionalCatchAllRouteSegments(left[limit:]) {
			return -1
		}
		return 1
	}
}

func routeSegmentSpecificity(segment routeSegment) int {
	switch segment.Kind {
	case frameworkrouter.SegmentStatic:
		return 4
	case frameworkrouter.SegmentDynamic:
		return 3
	case frameworkrouter.SegmentCatchAll:
		return 2
	case frameworkrouter.SegmentOptionalCatchAll:
		return 1
	default:
		return 0
	}
}

func remainingOnlyOptionalCatchAllRouteSegments(segments []routeSegment) bool {
	if len(segments) == 0 {
		return false
	}
	for _, segment := range segments {
		if segment.Kind != frameworkrouter.SegmentOptionalCatchAll {
			return false
		}
	}
	return true
}
