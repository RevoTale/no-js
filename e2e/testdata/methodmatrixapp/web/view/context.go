package view

import (
	"net/http"
	"net/url"
)

type Context struct{}

func (c *Context) ResolveRoot(r *http.Request) *url.URL {
	root, _ := url.Parse("https://methods.example.test")
	return root
}
