package resolvers

import (
	"context"
	"net/http"

	"blog/framework"
	"blog/framework/metagen"
	"blog/internal/web/appcore"
	"blog/internal/web/seo"
)

func (Resolver) MetaGenNoteParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params NoteParamSlugParams,
) (metagen.Metadata, error) {
	return seo.MetaGenNotePage(ctx, appCtx, r, params.Slug)
}

func (Resolver) ResolveNoteParamSlugPage(
	ctx context.Context,
	appCtx *appcore.Context,
	r *http.Request,
	params NoteParamSlugParams,
) (appcore.NotePageView, error) {
	return appcore.LoadNotePage(ctx, appCtx, r, framework.SlugParams{Slug: params.Slug})
}
