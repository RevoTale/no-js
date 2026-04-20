package author

import (
	"net/http"
	"strings"
	"time"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Feed(runtimeCtx framework.RuntimeContext[*runtime.Context], r *http.Request) (discovery.FeedDocument, error) {
	slug := strings.TrimPrefix(r.URL.Path, "/author/")
	slug = strings.TrimSuffix(slug, "/feed.xml")
	root := runtimeCtx.ResolveRoot(r)
	publishedAt := time.Date(2026, time.April, 20, 8, 0, 0, 0, time.UTC)

	return discovery.FeedDocument{
		Title:       "Feed for " + slug,
		Link:        runtime.AbsoluteURL(root, "/author/"+slug),
		Description: "Latest author updates",
		Language:    "en",
		SelfURL:     runtime.AbsoluteURL(root, "/author/"+slug+"/feed.xml"),
		Items: []discovery.FeedItem{
			{
				Title:       "Profile " + slug,
				Link:        runtime.AbsoluteURL(root, "/author/"+slug),
				GUID:        "author-" + slug,
				Description: "Author profile",
				PublishedAt: &publishedAt,
			},
		},
	}, nil
}
