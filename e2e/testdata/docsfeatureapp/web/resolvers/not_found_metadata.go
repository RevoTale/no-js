package resolvers

import (
	"example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/v2/framework"
	"github.com/RevoTale/no-js/v2/framework/metagen"
)

func (Resolver) MetaGenRootNotFound(
	_ framework.MetaContext[*view.Context],
	_ framework.NotFoundContext,
	_ RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Not Found"}, nil
}

func (Resolver) MetaGenAuthorParamSlugNotFound(
	_ framework.MetaContext[*view.Context],
	_ framework.NotFoundContext,
	_ AuthorParamSlugParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Unknown Author"}, nil
}
