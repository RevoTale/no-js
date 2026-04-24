package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/typedmodelsapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "typed model fixture"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Typed Models"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.RootPageView, error) {
	return view.RootPageView{Heading: "Typed root model"}, nil
}

func (Resolver) MetaGenGroupSiteMarketingLayout(
	meta framework.MetaContext[*view.Context],
	params GroupSiteMarketingParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Typed Marketing"}, nil
}

func (Resolver) ResolveGroupSiteMarketingLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupSiteMarketingParams,
) (view.MarketingLayoutView, error) {
	return view.MarketingLayoutView{Shell: "marketing-shell"}, nil
}

func (Resolver) MetaGenGroupSiteMarketingPage(
	meta framework.MetaContext[*view.Context],
	params GroupSiteMarketingParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Typed Marketing"}, nil
}

func (Resolver) ResolveGroupSiteMarketingPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupSiteMarketingParams,
) (view.MarketingPageView, error) {
	return view.MarketingPageView{Heading: "Typed marketing model"}, nil
}

func (Resolver) ResolveGroupSiteMarketingSlotPromoDefault(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupSiteMarketingSlotPromoParams,
) (view.PromoDefaultView, error) {
	return view.PromoDefaultView{Message: "promo-default-model"}, nil
}

func (Resolver) ResolveRootNotFound(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	params RootParams,
) (view.RootNotFoundView, error) {
	return view.RootNotFoundView{Message: "typed-not-found-model"}, nil
}
