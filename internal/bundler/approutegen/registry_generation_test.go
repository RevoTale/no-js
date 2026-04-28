package approutegen

import (
	"bytes"
	"testing"

	"github.com/RevoTale/no-js/framework/metagen"
	"github.com/RevoTale/no-js/internal/bundler/clientassets"
	"github.com/RevoTale/no-js/internal/projectlayout"
	"github.com/stretchr/testify/require"
)

func TestResolverNamespaceGenerationDeterministic(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
		},
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
		},
	}

	first, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		nil,
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.NoError(t, err)
	second, err := generateResolverNamespaceSource(
		projectlayout.ProjectLayout{AppModulePath: testAppModulePath},
		metas,
		nil,
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.NoError(t, err)
	require.True(t, bytes.Equal(first, second))
	require.Contains(t, string(first), "var _ RouteResolver = (*Resolver)(nil)")
}

func TestRegistryGenerationUsesSingleResolverNamespace(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
			Page:           templateDef{ModuleName: "r_page_author_param_slug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "route_resolvers \"example.com/app/web/resolvers\"")
	require.NotContains(t, text, "rr_")
	require.Contains(t, text, "func NewRouteResolvers() RouteResolvers")
	require.Contains(t, text, "return &route_resolvers.Resolver{}")
	require.NotContains(t, text, "type NotFoundViewResolver interface")
	require.Contains(
		t,
		text,
		"func NotFoundPage(resolvers RouteResolvers) func(runtime framework.RuntimeContext[*view.Context], "+
			"r *http.Request, notFound framework.NotFoundContext) (templ.Component, error)",
	)
	require.Contains(t, text, "framework.PageOnlyRouteHandler")
	require.NotContains(t, text, "PageAndLiveRouteHandler")
	require.NotContains(t, text, "/.live/")
	require.NotContains(t, text, "ParseRootLiveState")
	require.Contains(t, text, "return renderNotFoundPage(resolvers, runtime, r, notFound)")
	require.Contains(t, text, "notFoundMeta, err := resolvers.MetaGenRootNotFound")
	require.Contains(t, text, "view, err := resolvers.ResolveRootNotFound")
	require.Contains(t, text, "RootLayout: r_root_root.RootLayout")
	require.Contains(t, text, "MetaGenContextChain: []framework.PageMetaGenContext")
	require.NotContains(t, text, "ErrorPage:")
	require.NotContains(t, text, "r_error_root")
}

func TestRegistryGenerationEmitsClientAssets(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.RootPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}
	plan := clientassets.Plan{
		RouteAssets: map[string]metagen.ClientAssets{
			"": {
				Stylesheets:   []string{"routes/index.css"},
				ModuleScripts: []string{"routes/index.js"},
			},
		},
		NotFoundAssets: map[string]metagen.ClientAssets{
			"": {Stylesheets: []string{"routes/404.css"}},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{Kind: rootTemplate, RouteID: "", ModuleName: "r_root_root"},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
		plan,
	)
	require.NoError(t, err)

	text := string(registry)
	require.Contains(t, text, "ClientAssets: metagen.ClientAssets{")
	require.Contains(t, text, `"routes/index.css"`)
	require.Contains(t, text, `"routes/index.js"`)
	require.Contains(t, text, "finalizeNotFoundMetadata(requestContext(r), meta, notFoundClientAssets(routeID))")
	require.Contains(t, text, "return metagen.MergeManagedClientAssets(ctx, meta, assets)")
	require.Contains(t, text, "func notFoundClientAssets(routeID string) metagen.ClientAssets")
	require.Contains(t, text, `"routes/404.css"`)
}

func TestRegistryGenerationRequiresRootNotFoundTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	_, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing root 404")
}

func TestRegistryGenerationCombinesRootAndDynamicNotFoundCases(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "author/_param__slug",
			RouteName:      "AuthorParamSlug",
			ParamsTypeName: "AuthorParamSlugParams",
			Params:         []routeParamDef{{Name: "slug", FieldName: "Slug"}},
			PageViewType:   "view.AuthorPageView",
			Page:           templateDef{ModuleName: "r_page_author_param_slug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
			"author/_param__slug": {
				Kind:       notFoundTemplate,
				RouteID:    "author/_param__slug",
				ModuleName: "r_not_found_author_param_slug",
				ModelType:  "view.AuthorNotFoundView",
			},
		},
	)
	require.NoError(t, err)
	require.Contains(t, string(registry), `case "", "author/_param__slug":`)
}

func TestRegistryGenerationDoesNotRequireRootErrorTemplate(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "",
			RouteName:      "Root",
			ParamsTypeName: "RootParams",
			PageViewType:   "view.NotesPageView",
			Page:           templateDef{ModuleName: "r_page_root"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)
	require.NotContains(t, string(registry), "ErrorPage:")
}

func TestRegistryGenerationDoesNotWireErrorTemplateModules(t *testing.T) {
	metas := []routeMeta{
		{
			RouteID:        "author/_param__slug/note/_param__noteSlug",
			RouteName:      "AuthorParamSlugNoteParamNoteslug",
			ParamsTypeName: "AuthorParamSlugNoteParamNoteslugParams",
			Params: []routeParamDef{
				{Name: "slug", FieldName: "Slug"},
				{Name: "noteSlug", FieldName: "Noteslug"},
			},
			PageViewType: "view.NotePageView",
			Page:         templateDef{ModuleName: "r_page_author_param_slug_note_param_noteslug"},
		},
	}

	registry, err := generateRegistrySource(
		projectlayout.ProjectLayout{GeneratedImport: "web/generated", AppModulePath: testAppModulePath},
		zeroArgViewInspection(),
		metas,
		nil,
		nil,
		nil,
		templateDef{
			Kind:       rootTemplate,
			RouteID:    "",
			ModuleName: "r_root_root",
		},
		map[string]templateDef{},
		map[string]templateDef{
			"": {
				Kind:       notFoundTemplate,
				RouteID:    "",
				ModuleName: "r_not_found_root",
				ModelType:  "view.RootNotFoundView",
			},
		},
	)
	require.NoError(t, err)

	text := string(registry)
	require.NotContains(t, text, "ErrorPage:")
	require.NotContains(t, text, "r_error_")
}

func TestGenerateSourcePackageParamsFileUsesPublicRouteID(t *testing.T) {
	source, err := generateSourcePackageParamsFile(sourcePackageDef{
		InternalRouteID: "_group__marketing/posts/_param__slug",
		PublicRouteID:   "posts/_param__slug",
		ParamsTypeName:  "GroupMarketingPostsParamSlugParams",
		Params: []routeParamDef{
			{Name: "slug", FieldName: "Slug", Type: "string"},
		},
		Package: "r_source_group_marketing_posts_param_slug",
	})
	require.NoError(t, err)
	require.Contains(t, string(source), `router.MatchPathPattern("/posts/_param__slug", requestPath)`)
}
