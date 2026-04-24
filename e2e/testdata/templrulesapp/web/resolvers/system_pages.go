package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/templrulesapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveGroupLabDashboardLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupLabDashboardParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Dashboard Variant"}, nil
}

func (Resolver) ResolveGroupLabDashboardSlotInspectorDefault(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupLabDashboardSlotInspectorParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Dashboard Variant"}, nil
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
