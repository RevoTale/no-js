package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveGroupMarketingDashboardLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupMarketingDashboardParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Dashboard"}, nil
}

func (Resolver) ResolveGroupMarketingDashboardSlotAnalyticsDefault(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupMarketingDashboardSlotAnalyticsParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Dashboard"}, nil
}

func (Resolver) ResolveRootNotFound(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ framework.NotFoundContext,
	_ RootParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Not Found"}, nil
}

func (Resolver) ResolveAuthorParamSlugNotFound(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ framework.NotFoundContext,
	_ AuthorParamSlugParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Unknown Author"}, nil
}
