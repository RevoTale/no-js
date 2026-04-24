package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/routepagecssapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveRootNotFound(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ framework.NotFoundContext,
	_ RootParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Not Found"}, nil
}
