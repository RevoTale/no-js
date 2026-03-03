package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenTalesPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ TalesParams,
) (metagen.Metadata, error) {
	return seo.MetaGenTalesPage(ctx, appCtx, r)
}

func (Resolver) ResolveTalesPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ TalesParams,
) (appcore.NotesPageView, error) {
	return appcore.LoadNotesTalesPage(ctx, appCtx, r, framework.EmptyParams{})
}
