package resolvers

import (
	"context"
	"net/http"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*runtime.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{
		Description: "grouped namespace fixture app",
	}, nil
}

func (Resolver) MetaGenRootPage(meta framework.MetaContext[*runtime.Context], params RootParams) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Grouped Namespace Fixture"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params RootParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Grouped Namespace Fixture"}, nil
}

func (Resolver) MetaGenGroupBlogDiscoverLayout(
	meta framework.MetaContext[*runtime.Context],
	params GroupBlogDiscoverParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Blog Discover"}, nil
}

func (Resolver) MetaGenGroupBlogDiscoverNotesPage(
	meta framework.MetaContext[*runtime.Context],
	params GroupBlogDiscoverNotesParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Notes Discover"}, nil
}

func (Resolver) ResolveGroupBlogDiscoverNotesPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params GroupBlogDiscoverNotesParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Notes Discover"}, nil
}

func (Resolver) MetaGenGroupBlogDiscoverGroupEditorialLayout(
	meta framework.MetaContext[*runtime.Context],
	params GroupBlogDiscoverGroupEditorialParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Editorial Discover"}, nil
}

func (Resolver) MetaGenGroupBlogDiscoverGroupEditorialGuidesPage(
	meta framework.MetaContext[*runtime.Context],
	params GroupBlogDiscoverGroupEditorialGuidesParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Guides Discover"}, nil
}

func (Resolver) ResolveGroupBlogDiscoverGroupEditorialGuidesPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params GroupBlogDiscoverGroupEditorialGuidesParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Guides Discover"}, nil
}

func (Resolver) MetaGenGroupShopDiscoverLayout(
	meta framework.MetaContext[*runtime.Context],
	params GroupShopDiscoverParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Shop Discover"}, nil
}

func (Resolver) MetaGenGroupShopDiscoverTagsPage(
	meta framework.MetaContext[*runtime.Context],
	params GroupShopDiscoverTagsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Tags Discover"}, nil
}

func (Resolver) ResolveGroupShopDiscoverTagsPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params GroupShopDiscoverTagsParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Tags Discover"}, nil
}
