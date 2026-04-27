package approutegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"slices"
	"sort"

	"github.com/RevoTale/no-js/internal/bundler/clientassets"
	"github.com/RevoTale/no-js/internal/bundler/viewcontract"
	"github.com/RevoTale/no-js/internal/projectlayout"
)

func generateRegistrySource(
	paths projectlayout.ProjectLayout,
	viewHooks viewcontract.Inspection,
	metas []routeMeta,
	slotMetas []routeMeta,
	slotOwners map[string][]slotDef,
	methodRoutes []methodRouteDef,
	root templateDef,
	layouts map[string]templateDef,
	notFounds map[string]templateDef,
	clientAssetPlans ...clientassets.Plan,
) ([]byte, error) {
	clientAssetPlan := clientassets.Plan{}
	if len(clientAssetPlans) > 0 {
		clientAssetPlan = clientAssetPlans[0]
	}

	if root.SourcePath == "" && root.ModuleName == "" {
		return nil, errors.New("missing root template metadata")
	}
	if _, ok := notFounds[""]; !ok {
		return nil, errors.New("missing root 404 template metadata")
	}
	allMetas := append(slices.Clone(metas), slotMetas...)
	allLayouts := mergeTemplateMaps(layouts, slotLayoutsFromOwners(slotOwners))
	defaults := defaultsFromSlotOwners(slotOwners)
	contracts, err := buildRouteContracts(allMetas, allLayouts, defaults, notFounds)
	if err != nil {
		return nil, fmt.Errorf("build route contracts: %w", err)
	}
	contractsByID := contractsByRouteID(contracts)
	slotNames := slotNamesByRoute(slotOwners)

	importLines := []string{
		"\"context\"",
		"\"net/http\"",
		"\"strings\"",
		fmt.Sprintf("%q", frameworkModulePath+"/framework"),
		fmt.Sprintf("frameworki18n %q", frameworkModulePath+"/framework/i18n"),
		fmt.Sprintf("%q", frameworkModulePath+"/framework/metagen"),
		fmt.Sprintf("%q", frameworkModulePath+"/framework/router"),
		fmt.Sprintf("%q", viewImportPath(paths)),
		fmt.Sprintf("route_resolvers %q", resolversImportPath(paths)),
		"\"github.com/a-h/templ\"",
	}
	if len(slotOwners) > 0 {
		importLines = append(importLines, "\"sync\"")
	}
	if anyParamsUseSlice(contracts) {
		importLines = append(importLines, "\"slices\"")
	}

	moduleImports := make([]string, 0, len(metas)+len(layouts)+len(notFounds)+1)
	moduleImports = append(moduleImports, fmt.Sprintf(
		"%s %q",
		root.ModuleName,
		generatedModuleImportPath(paths, root.ModuleName),
	))
	for _, meta := range metas {
		moduleImports = append(moduleImports, fmt.Sprintf(
			"%s %q",
			meta.Page.ModuleName,
			generatedModuleImportPath(paths, meta.Page.ModuleName),
		))
	}
	for _, meta := range slotMetas {
		moduleImports = append(moduleImports, fmt.Sprintf(
			"%s %q",
			meta.Page.ModuleName,
			generatedModuleImportPath(paths, meta.Page.ModuleName),
		))
	}

	layoutKeys := make([]string, 0, len(layouts))
	for routeID := range layouts {
		layoutKeys = append(layoutKeys, routeID)
	}
	sort.Strings(layoutKeys)
	for _, routeID := range layoutKeys {
		layout := layouts[routeID]
		moduleImports = append(moduleImports, fmt.Sprintf(
			"%s %q",
			layout.ModuleName,
			generatedModuleImportPath(paths, layout.ModuleName),
		))
	}
	slotLayoutKeys := make([]string, 0, len(slotOwners))
	for owner := range slotOwners {
		slotLayoutKeys = append(slotLayoutKeys, owner)
	}
	sort.Strings(slotLayoutKeys)
	for _, owner := range slotLayoutKeys {
		for _, slot := range slotOwners[owner] {
			if slot.Default != nil {
				moduleImports = append(moduleImports, fmt.Sprintf(
					"%s %q",
					slot.Default.ModuleName,
					generatedModuleImportPath(paths, slot.Default.ModuleName),
				))
			}
			layoutKeys := make([]string, 0, len(slot.Layouts))
			for routeID := range slot.Layouts {
				layoutKeys = append(layoutKeys, routeID)
			}
			sort.Strings(layoutKeys)
			for _, routeID := range layoutKeys {
				layout := slot.Layouts[routeID]
				moduleImports = append(moduleImports, fmt.Sprintf(
					"%s %q",
					layout.ModuleName,
					generatedModuleImportPath(paths, layout.ModuleName),
				))
			}
		}
	}

	notFoundKeys := make([]string, 0, len(notFounds))
	for routeID := range notFounds {
		notFoundKeys = append(notFoundKeys, routeID)
	}
	sort.Strings(notFoundKeys)
	for _, routeID := range notFoundKeys {
		notFound := notFounds[routeID]
		moduleImports = append(moduleImports, fmt.Sprintf(
			"%s %q",
			notFound.ModuleName,
			generatedModuleImportPath(paths, notFound.ModuleName),
		))
	}
	for _, route := range methodRoutes {
		moduleImports = append(moduleImports, fmt.Sprintf(
			"%s %q",
			route.PackageAlias,
			generatedModuleImportPath(paths, route.PackageModule),
		))
	}

	moduleImports = dedupeSorted(moduleImports)
	importLines = append(importLines, moduleImports...)

	buffer := &bytes.Buffer{}
	buffer.WriteString(generatedGoHeader + "\n")
	buffer.WriteString("package gen\n\n")
	buffer.WriteString("import (\n")
	for _, line := range importLines {
		buffer.WriteString("\t" + line + "\n")
	}
	buffer.WriteString(")\n\n")

	buffer.WriteString("type RouteResolvers = route_resolvers.RouteResolver\n\n")
	for _, meta := range metas {
		writef(buffer, "type %s = route_resolvers.%s\n", meta.ParamsTypeName, meta.ParamsTypeName)
	}
	if len(metas) > 0 {
		buffer.WriteString("\n")
	}
	buffer.WriteString("func NewRouteResolvers() RouteResolvers {\n")
	buffer.WriteString("\treturn &route_resolvers.Resolver{}\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString(
		"func NotFoundPage(resolvers RouteResolvers) " +
			"func(appCtx *view.Context, r *http.Request, notFound framework.NotFoundContext) (templ.Component, error) {\n",
	)
	buffer.WriteString(
		"	return func(appCtx *view.Context, r *http.Request, " +
			"notFound framework.NotFoundContext) (templ.Component, error) {\n",
	)
	buffer.WriteString("		return renderNotFoundPage(resolvers, appCtx, r, notFound)\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func requestContext(r *http.Request) context.Context {\n")
	buffer.WriteString("	if r == nil {\n")
	buffer.WriteString("		return context.Background()\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	return r.Context()\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString(
		"func withNotFoundRequestInfo(r *http.Request, " +
			"notFound framework.NotFoundContext) *http.Request {\n",
	)
	buffer.WriteString("	if r == nil {\n")
	buffer.WriteString("		return nil\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	info, _ := frameworki18n.RequestInfoFromContext(r.Context())\n")
	buffer.WriteString("	info.Locale = strings.TrimSpace(notFound.Locale)\n")
	buffer.WriteString("	if strings.TrimSpace(info.OriginalPath) == \"\" {\n")
	buffer.WriteString("		info.OriginalPath = strings.TrimSpace(notFound.RequestPath)\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	if strings.TrimSpace(info.StrippedPath) == \"\" {\n")
	buffer.WriteString("		info.StrippedPath = notFoundStrippedPath(notFound)\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	return r.WithContext(frameworki18n.WithRequestInfo(r.Context(), info))\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func notFoundStrippedPath(notFound framework.NotFoundContext) string {\n")
	buffer.WriteString("	requestPath := strings.TrimSpace(notFound.RequestPath)\n")
	buffer.WriteString("	if requestPath == \"\" {\n")
	buffer.WriteString("		return \"/\"\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	locale := strings.TrimSpace(notFound.Locale)\n")
	buffer.WriteString("	if locale == \"\" {\n")
	buffer.WriteString("		return requestPath\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	prefix := \"/\" + locale\n")
	buffer.WriteString("	if requestPath == prefix {\n")
	buffer.WriteString("		return \"/\"\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	if strings.HasPrefix(requestPath, prefix+\"/\") {\n")
	buffer.WriteString("		return strings.TrimPrefix(requestPath, prefix)\n")
	buffer.WriteString("	}\n")
	buffer.WriteString("	return requestPath\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("func Handlers(resolvers RouteResolvers) []framework.RouteHandler[*view.Context] {\n")
	buffer.WriteString("\treturn []framework.RouteHandler[*view.Context]{\n")
	for _, meta := range metas {
		writef(
			buffer,
			"\t\tframework.PageOnlyRouteHandler[*view.Context, %s, %s]{\n",
			meta.ParamsTypeName,
			meta.PageViewType,
		)

		if err := writePageModule(
			buffer,
			meta,
			root,
			layouts,
			contractsByID,
			clientAssetPlan.RouteAssets,
		); err != nil {
			return nil, err
		}
		buffer.WriteString("\t\t},\n")
	}
	for _, route := range methodRoutes {
		writef(
			buffer,
			"\t\tframework.MethodOnlyRouteHandler[*view.Context, %s.%s]{\n",
			route.PackageAlias,
			route.ParamsTypeName,
		)
		writef(
			buffer,
			"\t\t\tRoute: framework.MethodRouteModule[*view.Context, %s.%s]{\n",
			route.PackageAlias,
			route.ParamsTypeName,
		)
		writef(buffer, "\t\t\t\tRouteID: %q,\n", route.InternalRouteID)
		writef(buffer, "\t\t\t\tPattern: %q,\n", routePattern(route.PublicRouteID))
		writef(buffer, "\t\t\t\tParseParams: %s.ParseParams,\n", route.PackageAlias)
		for _, method := range route.Methods {
			writef(buffer, "\t\t\t\t%s: %s.%s,\n", method, route.PackageAlias, method)
		}
		buffer.WriteString("\t\t\t},\n")
		buffer.WriteString("\t\t},\n")
	}
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n\n")

	if err := writeNotFoundPageFunc(
		buffer,
		root,
		layouts,
		notFounds,
		slotNames,
		contractsByID,
		clientAssetPlan.NotFoundAssets,
	); err != nil {
		return nil, err
	}

	for _, contract := range contracts {
		writeParseParamsFunc(buffer, contract)
	}
	if err := writeSlotResolveFuncs(buffer, slotOwners, slotNames, contractsByID); err != nil {
		return nil, err
	}
	for _, meta := range metas {
		if err := writeComposeFunc(buffer, meta, slotOwners, slotNames, layouts, contractsByID); err != nil {
			return nil, err
		}
	}

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format registry source: %w", err)
	}
	return formatted, nil
}

func anyParamsUseSlice(contracts []routeContractDef) bool {
	for _, contract := range contracts {
		for _, param := range contract.Params {
			if param.Type == "[]string" {
				return true
			}
		}
	}
	return false
}

func generateDiscoverySource(paths projectlayout.ProjectLayout, discovery discoveryConventions) ([]byte, error) {
	importLines := []string{
		fmt.Sprintf("%q", frameworkModulePath+"/framework"),
		fmt.Sprintf("%q", frameworkModulePath+"/framework/discovery"),
		fmt.Sprintf("%q", viewImportPath(paths)),
	}

	packageImports := make([]string, 0, 1+len(discovery.Sitemaps)+len(discovery.Feeds))
	if discovery.HasRobots {
		packageImports = append(
			packageImports,
			fmt.Sprintf("%s %q", discoveryImportAlias(""), routesImportPath(paths)),
		)
	}
	for _, convention := range discovery.Sitemaps {
		packageImports = append(
			packageImports,
			fmt.Sprintf("%s %q", convention.ImportAlias, routesImportPathForRouteID(paths, convention.RouteID)),
		)
	}
	for _, convention := range discovery.Feeds {
		packageImports = append(
			packageImports,
			fmt.Sprintf("%s %q", convention.ImportAlias, routesImportPathForRouteID(paths, convention.RouteID)),
		)
	}
	if len(packageImports) > 0 {
		importLines = append(importLines, dedupeSorted(packageImports)...)
	}

	buffer := &bytes.Buffer{}
	buffer.WriteString(generatedGoHeader + "\n")
	buffer.WriteString("package gen\n\n")
	buffer.WriteString("import (\n")
	for _, line := range dedupeSorted(importLines) {
		buffer.WriteString("\t" + line + "\n")
	}
	buffer.WriteString(")\n\n")

	buffer.WriteString("func DiscoveryBundle() *discovery.Bundle[*view.Context] {\n")
	if !discovery.HasRobots && len(discovery.Sitemaps) == 0 && len(discovery.Feeds) == 0 {
		buffer.WriteString("\treturn nil\n")
		buffer.WriteString("}\n")
		buffer.WriteString("\n")
		buffer.WriteString("func DiscoveryExactHandlers() []framework.RouteHandler[*view.Context] {\n")
		buffer.WriteString("\treturn nil\n")
		buffer.WriteString("}\n")
		formatted, err := format.Source(buffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("format discovery source: %w", err)
		}
		return formatted, nil
	}
	buffer.WriteString("\treturn &discovery.Bundle[*view.Context]{\n")
	if discovery.HasRobots {
		writef(buffer, "\t\tRobots: %s.Robots,\n", discoveryImportAlias(""))
	}
	if len(discovery.Sitemaps) > 0 {
		buffer.WriteString("\t\tSitemaps: []discovery.SitemapRoute[*view.Context]{\n")
		for _, convention := range discovery.Sitemaps {
			buffer.WriteString("\t\t\t{\n")
			writef(buffer, "\t\t\t\tRoutePattern: %q,\n", routePattern(convention.RouteID))
			if convention.HasSitemap {
				writef(buffer, "\t\t\t\tSitemap: %s.Sitemap,\n", convention.ImportAlias)
			}
			if convention.HasGenerateSitemaps {
				writef(buffer, "\t\t\t\tGenerateSitemaps: %s.GenerateSitemaps,\n", convention.ImportAlias)
			}
			if convention.HasSitemapChunk {
				writef(buffer, "\t\t\t\tSitemapChunk: %s.SitemapChunk,\n", convention.ImportAlias)
			}
			buffer.WriteString("\t\t\t},\n")
		}
		buffer.WriteString("\t\t},\n")
	}
	if len(discovery.Feeds) > 0 {
		buffer.WriteString("\t\tFeeds: []discovery.FeedRoute[*view.Context]{\n")
		for _, convention := range discovery.Feeds {
			buffer.WriteString("\t\t\t{\n")
			writef(buffer, "\t\t\t\tRoutePattern: %q,\n", routePattern(convention.RouteID))
			writef(buffer, "\t\t\t\tFeed: %s.Feed,\n", convention.ImportAlias)
			buffer.WriteString("\t\t\t},\n")
		}
		buffer.WriteString("\t\t},\n")
	}
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n")
	buffer.WriteString("\n")
	buffer.WriteString("func DiscoveryExactHandlers() []framework.RouteHandler[*view.Context] {\n")
	buffer.WriteString("\treturn discovery.ExactHandlers(DiscoveryBundle())\n")
	buffer.WriteString("}\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format discovery source: %w", err)
	}
	return formatted, nil
}

func generateBundleSource(paths projectlayout.ProjectLayout, viewHooks viewcontract.Inspection) ([]byte, error) {
	buffer := &bytes.Buffer{}
	buffer.WriteString(generatedGoHeader + "\n")
	buffer.WriteString("package gen\n\n")
	buffer.WriteString("import (\n")
	writef(buffer, "\t%q\n", frameworkModulePath+"/framework/httpserver")
	writef(buffer, "\t%q\n", frameworkModulePath+"/framework/i18n")
	writef(buffer, "\tmessages %q\n", generatedI18nMessagesImportPath(paths))
	writef(buffer, "\t%q\n", viewImportPath(paths))
	buffer.WriteString(")\n\n")

	buffer.WriteString("func Bundle(appContext *view.Context) httpserver.AppBundle[*view.Context] {\n")
	buffer.WriteString("\tvar i18nConfig *i18n.Config\n")
	buffer.WriteString("\tcfg := messages.Config()\n")
	buffer.WriteString("\tif len(cfg.Locales) > 0 {\n")
	buffer.WriteString("\t\ti18nConfig = &cfg\n")
	buffer.WriteString("\t}\n\n")
	buffer.WriteString("	resolvers := NewRouteResolvers()\n\n")
	buffer.WriteString("\treturn httpserver.AppBundle[*view.Context]{\n")
	buffer.WriteString("\t\tContext:                       appContext,\n")
	buffer.WriteString("\t\tExactHandlers:                 DiscoveryExactHandlers(),\n")
	buffer.WriteString("\t\tHandlers:                      Handlers(resolvers),\n")
	buffer.WriteString("\t\tDiscovery:                     DiscoveryBundle(),\n")
	buffer.WriteString("\t\tI18n:                          i18nConfig,\n")
	buffer.WriteString("\t\tResolveRoot:                   appContext.ResolveRoot,\n")
	buffer.WriteString("\t\tNotFoundPage:                  NotFoundPage(resolvers),\n")
	if paths.Assets.TemplCSS {
		buffer.WriteString("\t\tTemplCSSClasses:               TemplCSSClasses,\n")
	} else {
		buffer.WriteString("\t\tTemplCSSClasses:               nil,\n")
	}
	if viewHooks.HasStaticAssetBasePathHook {
		buffer.WriteString("\t\tOnStaticAssetBasePathResolved: view.SetStaticAssetBasePath,\n")
	} else {
		buffer.WriteString("\t\tOnStaticAssetBasePathResolved: nil,\n")
	}
	buffer.WriteString("\t}\n")
	buffer.WriteString("}\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format bundle source: %w", err)
	}
	return formatted, nil
}
