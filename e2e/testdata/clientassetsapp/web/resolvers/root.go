package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/clientassetsapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "Client assets fixture"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.RootPageView, error) {
	return view.RootPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets"},
		Heading:        "Client Assets Home",
	}, nil
}

func (Resolver) MetaGenAboutPage(
	meta framework.MetaContext[*view.Context],
	params AboutParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets About"}, nil
}

func (Resolver) ResolveAboutPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params AboutParams,
) (view.AboutPageView, error) {
	return view.AboutPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets About"},
		Heading:        "About Without Client Assets",
	}, nil
}

func (Resolver) MetaGenSectionLayout(
	meta framework.MetaContext[*view.Context],
	params SectionParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Section Layout"}, nil
}

func (Resolver) ResolveSectionLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params SectionParams,
) (view.SectionLayoutView, error) {
	return view.SectionLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Section"},
	}, nil
}

func (Resolver) MetaGenSectionPage(
	meta framework.MetaContext[*view.Context],
	params SectionParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Section"}, nil
}

func (Resolver) ResolveSectionPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params SectionParams,
) (view.SectionPageView, error) {
	return view.SectionPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Section"},
		Heading:        "Section With Layout Assets",
	}, nil
}

func (Resolver) MetaGenComplexPage(
	meta framework.MetaContext[*view.Context],
	params ComplexParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Complex Selectors"}, nil
}

func (Resolver) ResolveComplexPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params ComplexParams,
) (view.ComplexPageView, error) {
	return view.ComplexPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Complex"},
		Heading:        "Complex Client Asset Selectors",
	}, nil
}

func (Resolver) MetaGenSectionSummaryPage(
	meta framework.MetaContext[*view.Context],
	params SectionSummaryParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Section Summary"}, nil
}

func (Resolver) ResolveSectionSummaryPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params SectionSummaryParams,
) (view.SectionSummaryPageView, error) {
	return view.SectionSummaryPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Section Summary"},
		Heading:        "Section Summary With Shared Layout CSS",
	}, nil
}

func (Resolver) MetaGenSectionAdminLayout(
	meta framework.MetaContext[*view.Context],
	params SectionAdminParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Section Admin Layout"}, nil
}

func (Resolver) ResolveSectionAdminLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params SectionAdminParams,
) (view.SectionAdminLayoutView, error) {
	return view.SectionAdminLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Section Admin"},
	}, nil
}

func (Resolver) MetaGenSectionAdminPage(
	meta framework.MetaContext[*view.Context],
	params SectionAdminParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Section Admin"}, nil
}

func (Resolver) ResolveSectionAdminPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params SectionAdminParams,
) (view.SectionAdminPageView, error) {
	return view.SectionAdminPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Client Assets Section Admin"},
		Heading:        "Section Admin With Nested Layout CSS",
	}, nil
}
