package i18nkeygen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildKeyDefsDeterministicNames(t *testing.T) {
	t.Parallel()

	messages := []Message{
		{ID: "layout.channelsButton", Translation: "Channels"},
		{ID: "note.publishedPrefix", Translation: "published"},
		{ID: "composer.readOnly", Translation: "read only"},
		{ID: "markdown.code.copy", Translation: "copy"},
	}

	defs, err := BuildKeyDefs(messages)
	require.NoError(t, err)
	require.Len(t, defs, len(messages))

	expectedByID := map[string]string{
		"composer.readOnly":     "ComposerReadOnly",
		"layout.channelsButton": "LayoutChannelsButton",
		"markdown.code.copy":    "MarkdownCodeCopy",
		"note.publishedPrefix":  "NotePublishedPrefix",
	}
	for _, def := range defs {
		expectedName, ok := expectedByID[def.ID]
		require.True(t, ok)
		require.Equal(t, expectedName, def.Name)
	}
}

func TestBuildKeyDefsDetectsConstNameCollision(t *testing.T) {
	t.Parallel()

	_, err := BuildKeyDefs([]Message{
		{ID: "a.bC", Translation: "first"},
		{ID: "aB.c", Translation: "second"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "collide")
}

func TestBuildKeyDefsRejectsReservedGeneratedNames(t *testing.T) {
	t.Parallel()

	_, err := BuildKeyDefs([]Message{
		{ID: "keys", Translation: "reserved"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "reserved generated var")
}

func TestGenerateFromJSONStableOutput(t *testing.T) {
	t.Parallel()

	source := []byte(`[
		{"id":"z.b","translation":"Z"},
		{"id":"a.b","translation":"A"}
	]`)

	first, err := GenerateFromJSON("i18n", source)
	require.NoError(t, err)
	second, err := GenerateFromJSON("i18n", source)
	require.NoError(t, err)

	require.Equal(t, string(first), string(second))
	require.Contains(t, string(first), "type Key string")
	require.Contains(t, string(first), "var Keys = []Key")
	require.Contains(t, string(first), "frameworki18n.Context[Key]")
}

func TestGenerateFromJSONIncludesTypedArgsHelpers(t *testing.T) {
	t.Parallel()

	source := []byte(`[
		{
			"id":"seo.author.description",
			"translation":"Browse notes by {{.Author}}.",
			"args":[{"name":"Author","type":"string"}]
		}
	]`)

	generated, err := GenerateFromJSON("i18n", source)
	require.NoError(t, err)

	output := string(generated)
	require.Contains(t, output, "type SeoAuthorDescriptionArgs struct")
	require.Contains(t, output, "Author string")
	require.Contains(
		t,
		output,
		"func TSeoAuthorDescription(ctx frameworki18n.Context[Key], args SeoAuthorDescriptionArgs) string",
	)
	require.Contains(t, output, `"Author": args.Author`)
}
