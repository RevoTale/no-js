package resolvers

import (
	"context"
	"net/http"

	runtime "example.com/no-js-e2e/customruntimeapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*runtime.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "custom runtime fixture"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*runtime.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Custom Runtime"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params RootParams,
) (runtime.RootPageView, error) {
	return runtime.RootPageView{
		RootLayoutView: runtime.RootLayoutView{PageTitle: "Custom Runtime"},
		Heading:        "Custom Runtime",
	}, nil
}
