package view

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Context struct{}

var staticAssetBasePath string

func (c *Context) ResolveRoot(r *http.Request) *url.URL {
	root, _ := url.Parse("https://custom.example.test")
	return root
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
