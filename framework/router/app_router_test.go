package router

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestAppRouterMatch(t *testing.T) {
	router, err := NewAppRouter(fstest.MapFS{
		"app/layout.templ":               {Data: []byte("package web")},
		"app/notes/page.templ":           {Data: []byte("package web")},
		"app/note/[slug]/page.templ":     {Data: []byte("package web")},
		"app/author/[slug]/page.templ":   {Data: []byte("package web")},
		"app/author/[slug]/layout.templ": {Data: []byte("package web")},
		"app/author/settings/page.templ": {Data: []byte("package web")},
		"app/author/[slug]/live/page.templ": {
			Data: []byte("package web"),
		},
	}, "app")
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectedID  string
		expectedKey string
		expectedVal string
	}{
		{name: "notes", path: "/notes", expectedID: "notes"},
		{name: "notes trailing slash", path: "/notes/", expectedID: "notes"},
		{
			name:        "note wildcard",
			path:        "/note/hello-world",
			expectedID:  "note/[slug]",
			expectedKey: "slug",
			expectedVal: "hello-world",
		},
		{name: "author static precedence", path: "/author/settings", expectedID: "author/settings"},
		{
			name:        "author wildcard",
			path:        "/author/nina",
			expectedID:  "author/[slug]",
			expectedKey: "slug",
			expectedVal: "nina",
		},
		{
			name:        "author nested wildcard",
			path:        "/author/nina/live",
			expectedID:  "author/[slug]/live",
			expectedKey: "slug",
			expectedVal: "nina",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, ok := router.Match(tc.path)
			require.True(t, ok)
			require.Equal(t, tc.expectedID, match.ID)
			if tc.expectedKey == "" {
				return
			}

			value, ok := match.Param(tc.expectedKey)
			require.True(t, ok)
			require.Equal(t, tc.expectedVal, value)
		})
	}
}

func TestAppRouterConflict(t *testing.T) {
	_, err := NewAppRouter(fstest.MapFS{
		"app/author/[slug]/page.templ": {Data: []byte("package web")},
		"app/author/[id]/page.templ":   {Data: []byte("package web")},
	}, "app")
	require.Error(t, err)
}

func TestMatchPathPattern(t *testing.T) {
	params, ok := MatchPathPattern("/author/[slug]/live", "/author/nina/live")
	require.True(t, ok)
	require.Equal(t, "nina", params["slug"])

	if _, ok = MatchPathPattern("/author/[slug]/live", "/author/nina"); ok {
		require.FailNow(t, "expected mismatch for shorter path")
	}
}

func TestIsValidSlug(t *testing.T) {
	require.True(t, IsValidSlug("l-you"))
	require.False(t, IsValidSlug("bad slug"))
}
