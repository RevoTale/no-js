package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenAuthorParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params AuthorParamSlugParams,
) (metagen.Metadata, error) {
	return seo.MetaGenAuthorPage(ctx, appCtx, r, params.Slug)
}

func (Resolver) ResolveAuthorParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params AuthorParamSlugParams,
) (appcore.AuthorPageView, error) {
	return appcore.LoadAuthorPage(ctx, appCtx, r, framework.SlugParams{Slug: params.Slug})
}
