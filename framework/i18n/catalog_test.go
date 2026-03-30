package i18n

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestCatalogLocalize(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"messages/active.en.json": {
			Data: []byte(`[
				{"id":"hello","translation":"Hello"},
				{"id":"greet","translation":"Hello {{.Name}}"}
			]`),
		},
		"messages/active.uk.json": {
			Data: []byte(`[
				{"id":"hello","translation":"Привіт"}
			]`),
		},
	}

	catalog, err := LoadCatalog(fsys, []string{
		"messages/active.en.json",
		"messages/active.uk.json",
	}, "en")
	require.NoError(t, err)

	require.Equal(t, "Привіт", catalog.Localize("uk", "hello", nil, "Hello"))
	require.Equal(t, "Fallback", catalog.Localize("uk", "missing", nil, "Fallback"))
	require.Equal(t, "Hello Bob", catalog.Localize("en", "greet", map[string]any{"Name": "Bob"}, "Hello Bob"))
}
