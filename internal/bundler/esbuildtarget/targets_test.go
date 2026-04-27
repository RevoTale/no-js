package esbuildtarget

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestParseDefaultsToES2020(t *testing.T) {
	t.Parallel()

	target, engines, err := Parse(nil)
	require.NoError(t, err)
	require.Equal(t, api.ES2020, target)
	require.Empty(t, engines)
}

func TestParseKeepsBrowserEnginesInOrder(t *testing.T) {
	t.Parallel()

	target, engines, err := Parse([]string{"es2020", "chrome107", "firefox104", "safari16"})
	require.NoError(t, err)
	require.Equal(t, api.ES2020, target)
	require.Equal(t, []api.Engine{
		{Name: api.EngineChrome, Version: "107"},
		{Name: api.EngineFirefox, Version: "104"},
		{Name: api.EngineSafari, Version: "16"},
	}, engines)
}

func TestParseRejectsEmptyBrowserTarget(t *testing.T) {
	t.Parallel()

	_, _, err := Parse([]string{"es2020", " "})
	require.ErrorContains(t, err, "browser target cannot be empty")
}
