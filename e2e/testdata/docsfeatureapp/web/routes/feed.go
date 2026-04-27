package routes

import (
	"net/http"
	"time"

	"example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtimeCtx framework.RuntimeContext[*view.Context], r *http.Request) (discovery.FeedDocument, error) {
	root := runtimeCtx.ResolveRoot(r)
	publishedAt := time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC)

	return discovery.FeedDocument{
		Title:       "Docs Feed",
		Link:        view.AbsoluteURL(root, "/"),
		Description: "Latest docs updates",
		Language:    "en",
		SelfURL:     view.AbsoluteURL(root, "/feed.xml"),
		Items: []discovery.FeedItem{
			{
				Title:       "Ada",
				Link:        view.AbsoluteURL(root, "/author/ada"),
				GUID:        "author-ada",
				Description: "Ada author page",
				PublishedAt: &publishedAt,
			},
		},
	}, nil
}
