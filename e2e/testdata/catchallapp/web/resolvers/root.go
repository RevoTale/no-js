package resolvers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"example.com/no-js-e2e/catchallapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "catchall fixture"}, nil
}

func (Resolver) MetaGenDocsCatchAllPartsPage(
	meta framework.MetaContext[*view.Context],
	params DocsCatchAllPartsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs Catch-All"}, nil
}

func (Resolver) ResolveDocsCatchAllPartsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params DocsCatchAllPartsParams,
) (view.CatchAllPageView, error) {
	return view.CatchAllPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Docs Catch-All"},
		Joined:         strings.Join(params.Parts, "/"),
		Depth:          strconv.Itoa(len(params.Parts)),
	}, nil
}
