package routes

import (
	"net/http"
	"time"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtimeCtx framework.RuntimeContext[*runtime.Context], r *http.Request) (discovery.FeedDocument, error) {
	root := runtimeCtx.ResolveRoot(r)
	publishedAt := time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC)

	return discovery.FeedDocument{
		Title:       "Docs Feed",
		Link:        runtime.AbsoluteURL(root, "/"),
		Description: "Latest docs updates",
		Language:    "en",
		SelfURL:     runtime.AbsoluteURL(root, "/feed.xml"),
		Items: []discovery.FeedItem{
			{
				Title:       "Ada",
				Link:        runtime.AbsoluteURL(root, "/author/ada"),
				GUID:        "author-ada",
				Description: "Ada author page",
				PublishedAt: &publishedAt,
			},
		},
	}, nil
}
