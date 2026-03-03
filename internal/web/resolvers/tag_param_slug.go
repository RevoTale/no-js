package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenTagParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params TagParamSlugParams,
) (metagen.Metadata, error) {
	return seo.MetaGenTagPage(ctx, appCtx, r, params.Slug)
}

func (Resolver) ResolveTagParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params TagParamSlugParams,
) (appcore.NotesPageView, error) {
	return appcore.LoadTagPage(ctx, appCtx, r, framework.SlugParams{Slug: params.Slug})
}
