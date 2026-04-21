package routes

import (
	"net/http"

	runtime "example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Sitemap(runtimeCtx framework.RuntimeContext[*runtime.Context], r *http.Request) ([]discovery.SitemapEntry, error) {
	root := runtimeCtx.ResolveRoot(r)

	return []discovery.SitemapEntry{
		{
			URL: runtime.AbsoluteURL(root, "/"),
			Alternates: map[string]string{
				"en": runtime.AbsoluteURL(root, "/"),
				"de": runtime.AbsoluteURL(root, "/de"),
			},
		},
		{
			URL: runtime.AbsoluteURL(root, "/dashboard"),
		},
	}, nil
}

func GenerateSitemaps(
	runtimeCtx framework.RuntimeContext[*runtime.Context],
	r *http.Request,
) ([]discovery.SitemapID, error) {
	root := runtimeCtx.ResolveRoot(r)
	return []discovery.SitemapID{
		{
			ID:       "authors",
			Location: runtime.AbsoluteURL(root, "/sitemap/authors.xml"),
		},
	}, nil
}

func SitemapChunk(
	runtimeCtx framework.RuntimeContext[*runtime.Context],
	r *http.Request,
	id string,
) ([]discovery.SitemapEntry, error) {
	if id != "authors" {
		return nil, nil
	}

	root := runtimeCtx.ResolveRoot(r)
	return []discovery.SitemapEntry{
		{
			URL: runtime.AbsoluteURL(root, "/author/ada"),
			Alternates: map[string]string{
				"en": runtime.AbsoluteURL(root, "/author/ada"),
				"de": runtime.AbsoluteURL(root, "/de/author/ada"),
			},
		},
	}, nil
}
