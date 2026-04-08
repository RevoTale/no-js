package router

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestAppRouterMatch(t *testing.T) {
	router, err := NewAppRouter(fstest.MapFS{
		"app/layout.templ":                     {Data: []byte("package web")},
		"app/notes/page.templ":                 {Data: []byte("package web")},
		"app/note/_param__slug/page.templ":     {Data: []byte("package web")},
		"app/author/_param__slug/page.templ":   {Data: []byte("package web")},
		"app/author/_param__slug/layout.templ": {Data: []byte("package web")},
		"app/author/settings/page.templ":       {Data: []byte("package web")},
		"app/author/_param__slug/live/page.templ": {
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
			expectedID:  "note/_param__slug",
			expectedKey: "slug",
			expectedVal: "hello-world",
		},
		{name: "author static precedence", path: "/author/settings", expectedID: "author/settings"},
		{
			name:        "author wildcard",
			path:        "/author/nina",
			expectedID:  "author/_param__slug",
			expectedKey: "slug",
			expectedVal: "nina",
		},
		{
			name:        "author nested wildcard",
			path:        "/author/nina/live",
			expectedID:  "author/_param__slug/live",
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

			values, ok := match.Param(tc.expectedKey)
			require.True(t, ok)
			require.Equal(t, []string{tc.expectedVal}, values)
		})
	}
}

func TestAppRouterConflict(t *testing.T) {
	_, err := NewAppRouter(fstest.MapFS{
		"app/author/_param__slug/page.templ": {Data: []byte("package web")},
		"app/author/_param__id/page.templ":   {Data: []byte("package web")},
	}, "app")
	require.Error(t, err)
}

func TestMatchPathPattern(t *testing.T) {
	params, ok := MatchPathPattern("/author/_param__slug/live", "/author/nina/live")
	require.True(t, ok)
	require.Equal(t, []string{"nina"}, params["slug"])

	if _, ok = MatchPathPattern("/author/_param__slug/live", "/author/nina"); ok {
		require.FailNow(t, "expected mismatch for shorter path")
	}
}

func TestMatchPathPatternCatchAllAndOptionalCatchAll(t *testing.T) {
	params, ok := MatchPathPattern("/docs/_catchall__slug", "/docs/a/b")
	require.True(t, ok)
	require.Equal(t, []string{"a", "b"}, params["slug"])

	params, ok = MatchPathPattern("/docs/_optional_catchall__slug", "/docs")
	require.True(t, ok)
	_, exists := params["slug"]
	require.True(t, exists)
	require.Nil(t, params["slug"])
}

func TestParseDirectorySegmentsReservedNamespace(t *testing.T) {
	segments, err := ParseDirectorySegments(
		"_group__marketing/_slot__analytics/_param__slug/_catchall__parts/_optional_catchall__tail",
	)
	require.NoError(t, err)
	require.Len(t, segments, 5)
	require.Equal(t, SegmentGroup, segments[0].Kind)
	require.Equal(t, "marketing", segments[0].Name)
	require.Equal(t, SegmentSlot, segments[1].Kind)
	require.Equal(t, "analytics", segments[1].Name)
	require.Equal(t, SegmentDynamic, segments[2].Kind)
	require.Equal(t, "slug", segments[2].Name)
	require.Equal(t, SegmentCatchAll, segments[3].Kind)
	require.Equal(t, "parts", segments[3].Name)
	require.Equal(t, SegmentOptionalCatchAll, segments[4].Kind)
	require.Equal(t, "tail", segments[4].Name)
}

func TestParseDirectorySegmentsRejectsUnknownReservedNamespace(t *testing.T) {
	_, err := ParseDirectorySegments("_unknown__marketing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown reserved route segment")
}

func TestParseDirectorySegmentsRejectsShortReservedNamespace(t *testing.T) {
	_, err := ParseDirectorySegments("_g__marketing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown reserved route segment")
}
