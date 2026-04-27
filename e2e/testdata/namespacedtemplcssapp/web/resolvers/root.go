package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/namespacedtemplcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{
		Description: "Namespaced fixture layout metadata",
	}, nil
}

func (Resolver) MetaGenGroupMarketingDashboardLayout(
	meta framework.MetaContext[*view.Context],
	params GroupMarketingDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Title: "Marketing Dashboard",
	}, nil
}

func (Resolver) MetaGenGroupMarketingDashboardPage(
	meta framework.MetaContext[*view.Context],
	params GroupMarketingDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Title:       "Marketing Dashboard",
		Description: "Grouped route metadata",
	}, nil
}

func (Resolver) ResolveGroupMarketingDashboardPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupMarketingDashboardParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{PageTitle: "Marketing Dashboard"}, nil
}
