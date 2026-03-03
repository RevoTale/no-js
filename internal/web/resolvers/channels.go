package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenChannelsPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ ChannelsParams,
) (metagen.Metadata, error) {
	return seo.MetaGenChannelsPage(ctx, appCtx, r)
}

func (Resolver) ResolveChannelsPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	_ ChannelsParams,
) (appcore.NotesPageView, error) {
	return appcore.LoadChannelsPage(ctx, appCtx, r, framework.EmptyParams{})
}
