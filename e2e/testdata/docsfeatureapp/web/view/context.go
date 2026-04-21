package runtime

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync/atomic"

	i18n "example.com/no-js-e2e/docsfeatureapp/web/generated/i18n"
	messages "example.com/no-js-e2e/docsfeatureapp/web/generated/i18n/messages"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
)

type Context struct {
	root           *url.URL
	expensiveLoads atomic.Int32
}

var staticAssetBasePath string

func NewContext() *Context {
	root, _ := url.Parse("https://example.com/blog")
	return &Context{root: root}
}

func (c *Context) ResolveRoot(r *http.Request) *url.URL {
	if c == nil || c.root == nil {
		return nil
	}

	clone := *c.root
	if r != nil {
		if strings.TrimSpace(r.Host) != "" {
			clone.Host = strings.TrimSpace(r.Host)
		}
		if r.TLS != nil {
			clone.Scheme = "https"
		}
	}
	return &clone
}

func (c *Context) I18n(r *http.Request) frameworki18n.Context[i18n.Key] {
	return messages.NewContext(r, c.ResolveRoot(r))
}

func (c *Context) RegisterExpensiveLoad() int {
	return int(c.expensiveLoads.Add(1))
}

func (c *Context) ExpensiveLoadCount() int {
	return int(c.expensiveLoads.Load())
}

func SetStaticAssetBasePath(prefix string) {
	staticAssetBasePath = strings.TrimRight(strings.TrimSpace(prefix), "/")
}

func StaticAssetURL(assetPath string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(assetPath), "/")
	if trimmed == "" {
		return staticAssetBasePath
	}
	return path.Join(staticAssetBasePath, trimmed)
}

func AbsoluteURL(root *url.URL, pathValue string) string {
	if root == nil {
		return ""
	}

	clone := *root
	base := strings.TrimSuffix(strings.TrimSpace(clone.Path), "/")
	normalized := strings.TrimSpace(pathValue)
	if normalized == "" {
		normalized = "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	if normalized == "/" {
		if base == "" {
			clone.Path = "/"
		} else {
			clone.Path = base
		}
	} else {
		joined := path.Join(base, strings.TrimPrefix(normalized, "/"))
		if !strings.HasPrefix(joined, "/") {
			joined = "/" + joined
		}
		clone.Path = joined
	}

	clone.RawQuery = ""
	clone.Fragment = ""
	return clone.String()
}
