package view

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Context struct {
	stream *StreamState
}

type StreamState struct {
	continueRender chan struct{}
	releaseOnce    sync.Once
}

func NewContext() *Context {
	return &Context{stream: NewStreamState()}
}

func NewStreamState() *StreamState {
	return &StreamState{
		continueRender: make(chan struct{}),
	}
}

func (c *Context) ResolveRoot(r *http.Request) *url.URL {
	if r == nil {
		return nil
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   strings.TrimSpace(r.Host),
		Path:   "/",
	}
}

func (c *Context) StreamState() *StreamState {
	if c == nil {
		return nil
	}
	return c.stream
}

func (c *Context) ReleaseStream() {
	if c == nil || c.stream == nil {
		return
	}
	c.stream.Release()
}

func (s *StreamState) Wait() {
	if s == nil || s.continueRender == nil {
		return
	}
	<-s.continueRender
}

func (s *StreamState) Release() {
	if s == nil {
		return
	}

	s.releaseOnce.Do(func() {
		if s.continueRender != nil {
			close(s.continueRender)
		}
	})
}
