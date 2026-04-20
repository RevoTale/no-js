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
	return metagen.Metadata{Description: "optional catchall fixture"}, nil
}

func (Resolver) MetaGenLibraryOptionalCatchAllPartsPage(
	meta framework.MetaContext[*runtime.Context],
	params LibraryOptionalCatchAllPartsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Library Optional Catch-All"}, nil
}

func (Resolver) ResolveLibraryOptionalCatchAllPartsPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params LibraryOptionalCatchAllPartsParams,
) (runtime.OptionalCatchAllPageView, error) {
	joined := strings.Join(params.Parts, "/")
	if joined == "" {
		joined = "root"
	}

	return runtime.OptionalCatchAllPageView{
		RootLayoutView: runtime.RootLayoutView{PageTitle: "Library Optional Catch-All"},
		Joined:         joined,
		Depth:          strconv.Itoa(len(params.Parts)),
	}, nil
}
