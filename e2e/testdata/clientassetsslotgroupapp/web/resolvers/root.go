package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/clientassetsslotgroupapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Client Assets Slot Group"}, nil
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
