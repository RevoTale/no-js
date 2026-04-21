package resolvers

import (
	"context"
	"net/http"

	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	runtime "example.com/no-js-e2e/i18nprefixalwaysapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*runtime.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Publisher: "Prefix Always Fixture"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*runtime.Context],
	params RootParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{
		Title: i18n.TUiHomeHeading(tr),
	}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *runtime.Context,
	r *http.Request,
	params RootParams,
) (runtime.HomePageView, error) {
	tr := appCtx.I18n(r)
	return runtime.HomePageView{
		RootLayoutView:  runtime.RootLayoutView{PageTitle: i18n.TUiHomeHeading(tr)},
		Heading:         i18n.TUiHomeHeading(tr),
		Kicker:          i18n.TUiHomeKicker(tr),
		Locale:          tr.Locale(),
		SwitchToEnglish: tr.SwitchURL("en").String(),
		SwitchToGerman:  tr.SwitchURL("de").String(),
	}, nil
}

func (Resolver) MetaGenGreetParamNamePage(
	meta framework.MetaContext[*runtime.Context],
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
	appCtx *runtime.Context,
	r *http.Request,
	params GreetParamNameParams,
) (runtime.GreetPageView, error) {
	tr := appCtx.I18n(r)
	return runtime.GreetPageView{
		RootLayoutView: runtime.RootLayoutView{
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
