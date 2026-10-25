package resolvers

import (
	"example.com/no-js-e2e/methodmatrixapp/web/view"
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
