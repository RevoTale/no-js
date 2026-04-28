package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/notfoundmetadataapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenRootLayout(meta framework.MetaContext[*view.Context]) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "Root layout metadata"}, nil
}

func (Resolver) MetaGenRootPage(
	meta framework.MetaContext[*view.Context],
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Not Found Metadata Home"}, nil
}

func (Resolver) ResolveRootPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params RootParams,
) (view.PageView, error) {
	return view.PageView{Heading: "Not Found Metadata Home"}, nil
}

func (Resolver) MetaGenRootNotFound(
	meta framework.MetaContext[*view.Context],
	notFound framework.NotFoundContext,
	params RootParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{
		Title:       "Root 404 Metadata Title",
		Description: "Root 404 metadata description",
		OpenGraph:   &metagen.OpenGraph{Type: "website"},
	}, nil
}

func (Resolver) ResolveRootNotFound(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	params RootParams,
) (view.NotFoundView, error) {
	return view.NotFoundView{Heading: "Root missing"}, nil
}

func (Resolver) MetaGenDocsLayout(
	meta framework.MetaContext[*view.Context],
	params DocsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Description: "Docs layout metadata"}, nil
}

func (Resolver) ResolveDocsLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params DocsParams,
) (view.DocsLayoutView, error) {
	return view.DocsLayoutView{Section: "docs"}, nil
}

func (Resolver) MetaGenDocsPage(
	meta framework.MetaContext[*view.Context],
	params DocsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs Index"}, nil
}

func (Resolver) ResolveDocsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params DocsParams,
) (view.PageView, error) {
	return view.PageView{Heading: "Docs Index"}, nil
}

func (Resolver) MetaGenDocsFailPage(
	meta framework.MetaContext[*view.Context],
	params DocsFailParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs Fail"}, nil
}

func (Resolver) ResolveDocsFailPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params DocsFailParams,
) (view.PageView, error) {
	return view.PageView{}, framework.ErrNotFound
}

func (Resolver) MetaGenDocsNotFound(
	meta framework.MetaContext[*view.Context],
	notFound framework.NotFoundContext,
	params DocsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Docs 404 Metadata Title"}, nil
}

func (Resolver) ResolveDocsNotFound(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	notFound framework.NotFoundContext,
	params DocsParams,
) (view.NotFoundView, error) {
	return view.NotFoundView{Heading: "Docs missing"}, nil
}
