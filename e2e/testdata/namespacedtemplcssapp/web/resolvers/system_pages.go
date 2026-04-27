package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/namespacedtemplcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveGroupMarketingDashboardLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupMarketingDashboardParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Marketing Dashboard"}, nil
}

func (Resolver) ResolveGroupMarketingDashboardSlotAnalyticsDefault(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupMarketingDashboardSlotAnalyticsParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Marketing Dashboard"}, nil
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
