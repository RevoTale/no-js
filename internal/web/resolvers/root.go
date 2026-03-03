package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenRootPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ RootParams,
) (metagen.Metadata, error) {
	return seo.MetaGenRootPage(ctx, appCtx, r)
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ RootParams,
) (appcore.NotesPageView, error) {
	return appcore.LoadNotesPage(ctx, appCtx, r, framework.EmptyParams{})
}
