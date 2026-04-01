package site

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubResolver struct {
	canonical string
	resolved  string
}

func (resolver stubResolver) CanonicalURL() string {
	return resolver.canonical
}

func (resolver stubResolver) Resolve(*http.Request) string {
	return resolver.resolved
}

type stubURLResolver struct {
	stubResolver
	canonicalRoot *url.URL
	resolvedRoot  *url.URL
}

func (resolver stubURLResolver) CanonicalRoot() *url.URL {
	return resolver.canonicalRoot
}

func (resolver stubURLResolver) ResolveRoot(*http.Request) *url.URL {
	return resolver.resolvedRoot
}

func TestResolveRootURLReturnsRequestResolvedURLWhenPresent(t *testing.T) {
	t.Parallel()

	resolved := ResolveRootURL(stubResolver{
		canonical: "https://canonical.example",
		resolved:  "https://request.example",
	}, httptest.NewRequest(http.MethodGet, "https://request.example/", nil))

	require.Equal(t, "https://request.example", resolved)
}

func TestResolveRootURLFallsBackToCanonicalURL(t *testing.T) {
	t.Parallel()

	resolved := ResolveRootURL(stubResolver{
		canonical: "https://canonical.example",
	}, nil)

	require.Equal(t, "https://canonical.example", resolved)
}

func TestResolveRootParsesStringResolverOutput(t *testing.T) {
	t.Parallel()

	resolved := ResolveRoot(stubResolver{
		canonical: "https://canonical.example/blog?x=1#y",
	}, nil)

	require.NotNil(t, resolved)
	require.Equal(t, "https://canonical.example/blog", resolved.String())
}

func TestResolveRootUsesURLResolverWithoutReparse(t *testing.T) {
	t.Parallel()

	resolved := ResolveRoot(stubURLResolver{
		canonicalRoot: &url.URL{
			Scheme:   "https",
			Host:     "canonical.example",
			Path:     "/blog",
			RawQuery: "x=1",
			Fragment: "y",
		},
		resolvedRoot: &url.URL{
			Scheme: "https",
			Host:   "request.example",
			Path:   "/blog",
		},
	}, httptest.NewRequest(http.MethodGet, "https://request.example/", nil))

	require.NotNil(t, resolved)
	require.Equal(t, "https://request.example/blog", resolved.String())
}
