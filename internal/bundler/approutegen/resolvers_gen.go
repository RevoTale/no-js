package approutegen

import (
	"bytes"
	"fmt"
	"go/format"
	"slices"
	"sort"

	"github.com/RevoTale/no-js/internal/projectlayout"
)

func generateResolverNamespaceSource(
	paths projectlayout.ProjectLayout,
	metas []routeMeta,
	slotMetas []routeMeta,
	layouts map[string]templateDef,
	slotLayouts map[string]templateDef,
	defaults map[string]templateDef,
	notFounds map[string]templateDef,
) ([]byte, error) {
	allMetas := append(slices.Clone(metas), slotMetas...)
	allLayouts := mergeTemplateMaps(layouts, slotLayouts)
	contracts, err := buildRouteContracts(allMetas, allLayouts, defaults, notFounds)
	if err != nil {
		return nil, fmt.Errorf("build route contracts: %w", err)
	}
	contractsByID := contractsByRouteID(contracts)

	layoutRouteIDs := make([]string, 0, len(layouts))
	for routeID := range layouts {
		if routeID == "" {
			continue
		}
		layoutRouteIDs = append(layoutRouteIDs, routeID)
	}
	sort.Strings(layoutRouteIDs)
	layoutModelRouteIDs := sortedTemplateRouteIDs(allLayouts)
	defaultRouteIDs := sortedTemplateRouteIDs(defaults)
	notFoundRouteIDs := sortedTemplateRouteIDs(notFounds)

	buffer := &bytes.Buffer{}
	buffer.WriteString(generatedGoHeader + "\n")
	buffer.WriteString("package resolvers\n\n")
	buffer.WriteString("import (\n")
	buffer.WriteString("\t\"" + frameworkModulePath + "/framework\"\n")
	buffer.WriteString("\t\"" + frameworkModulePath + "/framework/metagen\"\n")
	buffer.WriteString("\t\"" + viewImportPath(paths) + "\"\n")
	buffer.WriteString("\t\"context\"\n")
	buffer.WriteString("\t\"net/http\"\n")
	buffer.WriteString(")\n\n")

	for _, contract := range contracts {
		writeContractParamsStruct(buffer, contract)
	}

	buffer.WriteString("type RouteResolver interface {\n")
	buffer.WriteString(
		"\tMetaGenRootLayout(meta framework.MetaContext[*view.Context]) " +
			"(metagen.Metadata, error)\n",
	)
	for _, routeID := range layoutRouteIDs {
		layout := layouts[routeID]
		contract, ok := contractsByID[routeID]
		if !ok {
			return nil, fmt.Errorf("missing route contract for layout route %q", routeID)
		}
		writef(
			buffer,
			"\t%s(meta framework.MetaContext[*view.Context], params %s) "+
				"(metagen.Metadata, error)\n",
			metaGenLayoutMethod(layout),
			contract.ParamsTypeName,
		)
	}
	for _, routeID := range layoutModelRouteIDs {
		layout := allLayouts[routeID]
		contract, ok := contractsByID[routeID]
		if !ok {
			return nil, fmt.Errorf("missing route contract for layout route %q", routeID)
		}
		writef(
			buffer,
			"\t%s(ctx context.Context, appCtx *view.Context, r *http.Request, params %s) (%s, error)\n",
			resolveLayoutMethod(layout),
			contract.ParamsTypeName,
			layout.ModelType,
		)
	}
	for _, routeID := range defaultRouteIDs {
		fallback := defaults[routeID]
		contract, ok := contractsByID[routeID]
		if !ok {
			return nil, fmt.Errorf("missing route contract for default route %q", routeID)
		}
		writef(
			buffer,
			"\t%s(ctx context.Context, appCtx *view.Context, r *http.Request, params %s) (%s, error)\n",
			resolveDefaultMethod(fallback),
			contract.ParamsTypeName,
			fallback.ModelType,
		)
	}
	for _, routeID := range notFoundRouteIDs {
		notFound := notFounds[routeID]
		contract, ok := contractsByID[routeID]
		if !ok {
			return nil, fmt.Errorf("missing route contract for not-found route %q", routeID)
		}
		writef(
			buffer,
			"\t%s(ctx context.Context, appCtx *view.Context, r *http.Request, "+
				"notFound framework.NotFoundContext, params %s) (%s, error)\n",
			resolveNotFoundMethod(notFound),
			contract.ParamsTypeName,
			notFound.ModelType,
		)
	}
	for _, meta := range metas {
		writef(
			buffer,
			"\t%s(meta framework.MetaContext[*view.Context], params %s) "+
				"(metagen.Metadata, error)\n",
			metaGenPageMethod(meta),
			meta.ParamsTypeName,
		)
	}
	for _, meta := range metas {
		writef(
			buffer,
			"\t%s(ctx context.Context, appCtx *view.Context, r *http.Request, params %s) (%s, error)\n",
			resolvePageMethod(meta),
			meta.ParamsTypeName,
			meta.PageViewType,
		)
	}
	for _, meta := range slotMetas {
		writef(
			buffer,
			"\t%s(ctx context.Context, appCtx *view.Context, r *http.Request, params %s) (%s, error)\n",
			resolvePageMethod(meta),
			meta.ParamsTypeName,
			meta.PageViewType,
		)
	}
	buffer.WriteString("}\n\n")
	buffer.WriteString("type Resolver struct{}\n\n")
	buffer.WriteString("var _ RouteResolver = (*Resolver)(nil)\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format resolver namespace source: %w", err)
	}
	return formatted, nil
}
