package view

import (
	"net/http"
	"net/url"
)

type Context struct{}

func (c *Context) ResolveRoot(*http.Request) *url.URL {
	root, _ := url.Parse("https://not-found.example.test")
	return root
}
