package routes

import (
	"net/http"

	"example.com/no-js-e2e/docsfeatureapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Robots(runtimeCtx framework.RuntimeContext[*view.Context], r *http.Request) (discovery.Robots, error) {
	root := runtimeCtx.ResolveRoot(r)

	return discovery.Robots{
		Rules: []discovery.RobotsRule{
			{
				UserAgent: "*",
				Allow:     []string{"/"},
				Disallow:  []string{"/api"},
			},
		},
		Sitemaps: []string{
			view.AbsoluteURL(root, "/sitemap-index.xml"),
		},
		Host: root.Host,
	}, nil
}
