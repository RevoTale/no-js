package resolvers

import (
	"context"
	"net/http"

	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	"example.com/no-js-e2e/i18nprefixalwaysapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Publisher: "Prefix Always Fixture"}, nil
}

func (Resolver) ResolveNotFoundView(
	_ context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
) view.RootLayoutView {
	tr := appCtx.I18n(r)
	return view.RootLayoutView{
		PageTitle:         i18n.TUiNotFoundTitle(tr),
		NotFoundHeading:   i18n.TUiNotFoundHeading(tr),
		SystemLocale:      tr.Locale(),
		SystemRequestPath: notFound.RequestPath,
		NotFoundSource:    string(notFound.Source),
	}
}

func (Resolver) ResolveFailPage(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ FailParams,
) (view.FailPageView, error) {
	return view.FailPageView{}, framework.ErrNotFound
}

func (Resolver) MetaGenFailPage(
	meta framework.MetaContext[*view.Context],
	params FailParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{
		Title: i18n.TUiErrorTitle(tr),
	}, nil
}

func (Resolver) MetaGenGroupSupportHelpLayout(
	meta framework.MetaContext[*view.Context],
	params GroupSupportHelpParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Support Help"}, nil
}

func (Resolver) MetaGenGroupSupportHelpFailPage(
	meta framework.MetaContext[*view.Context],
	params GroupSupportHelpFailParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{
		Title: i18n.TUiNotFoundTitle(tr),
	}, nil
}

func (Resolver) ResolveGroupSupportHelpFailPage(
	_ context.Context,
	_ *view.Context,
	_ *http.Request,
	_ GroupSupportHelpFailParams,
) (view.RootLayoutView, error) {
	return view.RootLayoutView{}, framework.ErrNotFound
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{
		Title: i18n.TUiHomeHeading(tr),
	}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.HomePageView, error) {
	tr := appCtx.I18n(r)
	return view.HomePageView{
		RootLayoutView:  view.RootLayoutView{PageTitle: i18n.TUiHomeHeading(tr)},
		Heading:         i18n.TUiHomeHeading(tr),
		Kicker:          i18n.TUiHomeKicker(tr),
		Locale:          tr.Locale(),
		SwitchToEnglish: tr.SwitchURL("en").String(),
		SwitchToGerman:  tr.SwitchURL("de").String(),
	}, nil
}

func (Resolver) MetaGenGreetParamNamePage(
	meta framework.MetaContext[*view.Context],
	params GreetParamNameParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	alternates, err := meta.Alternates(meta.Locale(), nil)
	if err != nil {
		return metagen.Metadata{}, err
	}

	return metagen.Metadata{
		Title: i18n.TUiGreetHeading(tr, i18n.UiGreetHeadingArgs{Name: params.Name}),
		Description: i18n.TSeoGreetDescription(tr, i18n.SeoGreetDescriptionArgs{
			Name: params.Name,
		}),
		Alternates: alternates,
	}, nil
}

func (Resolver) ResolveGreetParamNamePage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GreetParamNameParams,
) (view.GreetPageView, error) {
	tr := appCtx.I18n(r)
	return view.GreetPageView{
		RootLayoutView: view.RootLayoutView{
			PageTitle: i18n.TUiGreetHeading(tr, i18n.UiGreetHeadingArgs{Name: params.Name}),
		},
		Heading: i18n.TUiGreetHeading(tr, i18n.UiGreetHeadingArgs{
			Name: params.Name,
		}),
		Description: i18n.TSeoGreetDescription(tr, i18n.SeoGreetDescriptionArgs{
			Name: params.Name,
		}),
		Locale:          tr.Locale(),
		SwitchToEnglish: tr.SwitchURL("en").String(),
		SwitchToGerman:  tr.SwitchURL("de").String(),
	}, nil
}
