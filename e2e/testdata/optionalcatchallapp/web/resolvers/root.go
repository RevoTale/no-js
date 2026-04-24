package resolvers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"example.com/no-js-e2e/optionalcatchallapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "optional catchall fixture"}, nil
}

func (Resolver) MetaGenLibraryOptionalCatchAllPartsPage(
	meta framework.MetaContext[*view.Context],
	params LibraryOptionalCatchAllPartsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Library Optional Catch-All"}, nil
}

func (Resolver) ResolveLibraryOptionalCatchAllPartsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params LibraryOptionalCatchAllPartsParams,
) (view.OptionalCatchAllPageView, error) {
	joined := strings.Join(params.Parts, "/")
	if joined == "" {
		joined = "root"
	}

	return view.OptionalCatchAllPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Library Optional Catch-All"},
		Joined:         joined,
		Depth:          strconv.Itoa(len(params.Parts)),
	}, nil
}
