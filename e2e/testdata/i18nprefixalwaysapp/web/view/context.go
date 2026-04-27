package view

import (
	"net/http"
	"net/url"
	"strings"

	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	messages "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n/messages"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
)

type Context struct {
	root *url.URL
}

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
		if r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
			clone.Scheme = "https"
		}
	}
	return &clone
}

func (c *Context) I18n(r *http.Request) frameworki18n.Context[i18n.Key] {
	return messages.NewContext(r, c.ResolveRoot(r))
}
