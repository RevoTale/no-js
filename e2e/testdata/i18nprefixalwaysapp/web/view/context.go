package runtime

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	i18n "example.com/templcssapp/web/generated/i18n"
	messages "example.com/templcssapp/web/generated/i18n/messages"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
)

type Context struct {
	root *url.URL
}

var staticAssetBasePath string

func NewContext() *Context {
	root, _ := url.Parse("https://prefix.example.test")
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
