package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/routepagecssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "Route page CSS fixture"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Route Page CSS"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.RootPageView, error) {
	return view.RootPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Route Page CSS"},
		Heading:        "Route CSS Home",
	}, nil
}
