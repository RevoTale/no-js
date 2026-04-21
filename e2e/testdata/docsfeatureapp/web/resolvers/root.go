package resolvers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	i18n "example.com/no-js-e2e/docsfeatureapp/web/generated/i18n"
	runtime "example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*runtime.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{
		Publisher: "Docs Fixture Publisher",
	}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*runtime.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs Home"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params RootParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Docs Home"}, nil
}

func (Resolver) MetaGenGroupMarketingDashboardLayout(
	meta framework.MetaContext[*runtime.Context],
	params GroupMarketingDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Dashboard"}, nil
}

func (Resolver) MetaGenGroupMarketingDashboardPage(
	meta framework.MetaContext[*runtime.Context],
	params GroupMarketingDashboardParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Dashboard"}, nil
}

func (Resolver) ResolveGroupMarketingDashboardPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params GroupMarketingDashboardParams,
) (runtime.RootLayoutView, error) {
	return runtime.RootLayoutView{PageTitle: "Dashboard"}, nil
}

func (Resolver) MetaGenAuthorParamSlugPage(
	meta framework.MetaContext[*runtime.Context],
	params AuthorParamSlugParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	canonical := meta.LocalizedURL(meta.Locale(), "/author/"+params.Slug)
	alternates, err := meta.Alternates(meta.Locale(), nil)
	if err != nil {
		return metagen.Metadata{}, err
	}

	return metagen.Metadata{
		Title:       i18n.TUiAuthorHeading(tr, i18n.UiAuthorHeadingArgs{Author: params.Slug}),
		Description: i18n.TSeoAuthorDescription(tr, i18n.SeoAuthorDescriptionArgs{Author: params.Slug}),
		Alternates:  alternates,
		OpenGraph: &metagen.OpenGraph{
			Type: "profile",
			URL:  canonical.String(),
		},
	}, nil
}

func (Resolver) ResolveAuthorParamSlugPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params AuthorParamSlugParams,
) (runtime.AuthorPageView, error) {
	switch params.Slug {
	case "missing":
		return runtime.AuthorPageView{}, framework.ErrNotFound
	case "boom":
		return runtime.AuthorPageView{}, errors.New("boom")
	}

	tr := appCtx.I18n(r)
	cacheKey := "author:" + params.Slug
	loadShared := func(runCtx context.Context) (string, error) {
		count := appCtx.RegisterExpensiveLoad()
		return fmt.Sprintf("cached:%s:%d", params.Slug, count), nil
	}

	first, err := framework.CachedCall(ctx, cacheKey, loadShared)
	if err != nil {
		return runtime.AuthorPageView{}, err
	}
	second, err := framework.CachedCall(ctx, cacheKey, loadShared)
	if err != nil {
		return runtime.AuthorPageView{}, err
	}

	return runtime.AuthorPageView{
		RootLayoutView: runtime.RootLayoutView{
			PageTitle: i18n.TUiAuthorHeading(tr, i18n.UiAuthorHeadingArgs{Author: params.Slug}),
		},
		Heading:         i18n.TUiAuthorHeading(tr, i18n.UiAuthorHeadingArgs{Author: params.Slug}),
		Description:     i18n.TSeoAuthorDescription(tr, i18n.SeoAuthorDescriptionArgs{Author: params.Slug}),
		SharedA:         first,
		SharedB:         second,
		SwitchToEnglish: tr.SwitchURL("en").String(),
		LoadCount:       strconv.Itoa(appCtx.ExpensiveLoadCount()),
		Progress:        72,
	}, nil
}
