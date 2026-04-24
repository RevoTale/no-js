package routes

import (
	"net/http"

	"example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtimeCtx framework.RuntimeContext[*view.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	root := runtimeCtx.ResolveRoot(r)

	return []discovery.SitemapEntry{
		{
			URL: view.AbsoluteURL(root, "/"),
			Alternates: map[string]string{
				"en": view.AbsoluteURL(root, "/"),
				"de": view.AbsoluteURL(root, "/de"),
			},
		},
		{
			URL: view.AbsoluteURL(root, "/dashboard"),
		},
	}, nil
}

func GenerateSitemaps(
	runtimeCtx framework.RuntimeContext[*view.Context],
	r *http.Request,
) ([]discovery.SitemapID, error) {
	root := runtimeCtx.ResolveRoot(r)
	return []discovery.SitemapID{
		{
			ID:       "authors",
			Location: view.AbsoluteURL(root, "/sitemap/authors.xml"),
		},
	}, nil
}

func SitemapChunk(
	runtimeCtx framework.RuntimeContext[*view.Context],
	r *http.Request,
	id string,
) ([]discovery.SitemapEntry, error) {
	if id != "authors" {
		return nil, nil
	}

	root := runtimeCtx.ResolveRoot(r)
	return []discovery.SitemapEntry{
		{
			URL: view.AbsoluteURL(root, "/author/ada"),
			Alternates: map[string]string{
				"en": view.AbsoluteURL(root, "/author/ada"),
				"de": view.AbsoluteURL(root, "/de/author/ada"),
			},
		},
	}, nil
}
