package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/groupednamespaceapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveGroupBlogDiscoverLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupBlogDiscoverParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Blog Discover"}, nil
}

func (Resolver) ResolveGroupBlogDiscoverGroupEditorialLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupBlogDiscoverGroupEditorialParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Editorial Discover"}, nil
}

func (Resolver) ResolveGroupShopDiscoverLayout(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupShopDiscoverParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Shop Discover"}, nil
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
