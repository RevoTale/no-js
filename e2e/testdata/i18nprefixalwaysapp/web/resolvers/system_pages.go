package resolvers

import (
	"context"
	"net/http"

	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	"example.com/no-js-e2e/i18nprefixalwaysapp/web/view"
	"github.com/RevoTale/no-js/framework"
)

func (Resolver) ResolveGroupSupportHelpLayout(
	_ context.Context,
	appCtx *view.Context,
	r *http.Request,
	_ GroupSupportHelpParams,
) (view.RootLayoutView, error) {
	tr := appCtx.I18n(r)
	return view.RootLayoutView{
		PageTitle:    i18n.TUiNotFoundTitle(tr),
		SystemLocale: tr.Locale(),
	}, nil
}

func (Resolver) ResolveRootNotFound(
	_ context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	_ RootParams,
) (view.RootLayoutView, error) {
	return localizedNotFoundView(appCtx, r, notFound), nil
}

func (Resolver) ResolveGroupSupportHelpNotFound(
	_ context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	_ GroupSupportHelpParams,
) (view.RootLayoutView, error) {
	return localizedNotFoundView(appCtx, r, notFound), nil
}

func localizedNotFoundView(
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
