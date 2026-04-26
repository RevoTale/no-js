package clientassetext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScriptExtensionsIncludeTSX(t *testing.T) {
	require.True(t, IsScriptExtension(".tsx"))
	require.True(t, IsScriptFile("web/routes/page.tsx"))
	require.Equal(t, "page.js, page.ts, page.tsx, page.mjs, or page.mts", ScriptChoices("page"))
	require.Equal(t, "page.js", ScriptOutputName("page"))
}

func TestGeneratedHelpers(t *testing.T) {
	require.True(t, IsGeneratedHelperName("page.tsx_gen.go"))
	require.True(t, IsGeneratedHelperFor("page", "page.tsx_gen.go"))

	stem, ok := GeneratedHelperStem("page.tsx_gen.go")
	require.True(t, ok)
	require.Equal(t, "page", stem)
}
