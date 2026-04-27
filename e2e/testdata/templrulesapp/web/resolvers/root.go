package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/templrulesapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{
		Description: "templ rules fixture app",
	}, nil
}

func (Resolver) MetaGenRootPage(meta framework.MetaContext[*view.Context], params RootParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Templ Rules Fixture"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Templ Rules Fixture"}, nil
}

func (Resolver) MetaGenCardPage(meta framework.MetaContext[*view.Context], params CardParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Card Variant"}, nil
}

func (Resolver) ResolveCardPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params CardParams,
) (view.CardPageView, error) {
	return view.CardPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Card Variant"},
		Title:          "Urgent Card",
		Urgent:         true,
	}, nil
}

func (Resolver) MetaGenPanelPage(meta framework.MetaContext[*view.Context], params PanelParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Panel Variant"}, nil
}

func (Resolver) ResolvePanelPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params PanelParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Panel Variant"}, nil
}

func (Resolver) MetaGenBoardPage(meta framework.MetaContext[*view.Context], params BoardParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Board Variant"}, nil
}

func (Resolver) ResolveBoardPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params BoardParams,
) (view.BoardPageView, error) {
	return view.BoardPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Board Variant"},
		Title:          "Board Header",
		Body:           "Board Body",
	}, nil
}

func (Resolver) MetaGenDepsPage(meta framework.MetaContext[*view.Context], params DepsParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Deps Variant"}, nil
}

func (Resolver) ResolveDepsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params DepsParams,
) (view.MeterPageView, error) {
	return view.MeterPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Deps Variant"},
		Progress:       "72%",
		Accent:         "#17324d",
	}, nil
}

func (Resolver) MetaGenHooksPage(meta framework.MetaContext[*view.Context], params HooksParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Hooks Variant"}, nil
}

func (Resolver) ResolveHooksPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params HooksParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Hooks Variant"}, nil
}

func (Resolver) MetaGenVarsPage(meta framework.MetaContext[*view.Context], params VarsParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Vars Variant"}, nil
}

func (Resolver) ResolveVarsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params VarsParams,
) (view.MeterPageView, error) {
	return view.MeterPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Vars Variant"},
		Progress:       "72%",
		Accent:         "#17324d",
	}, nil
}

func (Resolver) MetaGenFallbackPage(
	meta framework.MetaContext[*view.Context],
	params FallbackParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Fallback Variant"}, nil
}

func (Resolver) ResolveFallbackPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params FallbackParams,
) (view.ProgressPageView, error) {
	return view.ProgressPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Fallback Variant"},
		Percent:        33,
	}, nil
}

func (Resolver) MetaGenMetadataPage(
	meta framework.MetaContext[*view.Context],
	params MetadataParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Title:       "Metadata Variant",
		Description: "Managed metadata from metagen",
		Alternates: metagen.Alternates{
			Canonical: meta.URL("/metadata").String(),
			Types: map[string]string{
				"application/json": meta.URL("/metadata.json").String(),
			},
		},
		Authors: []metagen.Author{
			{
				Name: "Fixture Author",
				URL:  meta.URL("/authors/fixture").String(),
			},
		},
		Publisher:     "Fixture Publisher",
		DangerRawHead: []string{`<meta name="fixture" content="metadata-variant">`},
	}, nil
}

func (Resolver) ResolveMetadataPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params MetadataParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Metadata Variant"}, nil
}

func (Resolver) MetaGenGroupLabDashboardLayout(
	meta framework.MetaContext[*view.Context],
	params GroupLabDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Dashboard Variant"}, nil
}

func (Resolver) MetaGenGroupLabDashboardPage(
	meta framework.MetaContext[*view.Context],
	params GroupLabDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Dashboard Variant"}, nil
}

func (Resolver) ResolveGroupLabDashboardPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabDashboardParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Dashboard Variant"}, nil
}
