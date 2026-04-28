package resolvers

import (
	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	"example.com/no-js-e2e/i18nprefixalwaysapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootNotFound(
	meta framework.MetaContext[*view.Context],
	_ framework.NotFoundContext,
	_ RootParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{Title: i18n.TUiNotFoundTitle(tr)}, nil
}

func (Resolver) MetaGenGroupSupportHelpNotFound(
	meta framework.MetaContext[*view.Context],
	_ framework.NotFoundContext,
	_ GroupSupportHelpParams,
) (metagen.Metadata, error) {
	tr := meta.App().I18n(meta.Request())
	return metagen.Metadata{Title: i18n.TUiNotFoundTitle(tr)}, nil
}
