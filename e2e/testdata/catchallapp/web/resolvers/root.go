package resolvers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*runtime.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "catchall fixture"}, nil
}

func (Resolver) MetaGenDocsCatchAllPartsPage(
	meta framework.MetaContext[*runtime.Context],
	params DocsCatchAllPartsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs Catch-All"}, nil
}

func (Resolver) ResolveDocsCatchAllPartsPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params DocsCatchAllPartsParams,
) (runtime.CatchAllPageView, error) {
	return runtime.CatchAllPageView{
		RootLayoutView: runtime.RootLayoutView{PageTitle: "Docs Catch-All"},
		Joined:         strings.Join(params.Parts, "/"),
		Depth:          strconv.Itoa(len(params.Parts)),
	}, nil
}
