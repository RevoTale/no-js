package i18n

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

type localeEntry struct {
	ID          string       `json:"id"`
	Translation string       `json:"translation"`
	Args        []MessageArg `json:"args,omitempty"`
}

func TestDiscoverMessageFilesSorted(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/layout.de.json": &fstest.MapFile{Data: []byte("[]")},
		"messages/seo.en.json":    &fstest.MapFile{Data: []byte("[]")},
		"messages/notes.en.json":  &fstest.MapFile{Data: []byte("[]")},
	}

	files, err := DiscoverMessageFiles(filesystem)
	require.NoError(t, err)
	require.Equal(t, "messages/layout.de.json,messages/notes.en.json,messages/seo.en.json", strings.Join(files, ","))
}

func TestDiscoverMessageFilesRejectsSubdirectories(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/nested/active.en.json": &fstest.MapFile{Data: []byte("[]")},
	}

	_, err := DiscoverMessageFiles(filesystem)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must not contain subdirectories")
}

func TestDiscoverMessageFilesRejectsNonJSON(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: []byte("[]")},
		"messages/README.md":      &fstest.MapFile{Data: []byte("docs")},
	}

	_, err := DiscoverMessageFiles(filesystem)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must contain only json files")
}

func TestDiscoverMessageFilesRejectsInvalidLocaleSuffix(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/layout.english.json": &fstest.MapFile{Data: []byte("[]")},
	}

	_, err := DiscoverMessageFiles(filesystem)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must end with .<locale>.json")
}

func TestLoadMessageDefinitionsMergesLocaleShards(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/layout.en.json": &fstest.MapFile{Data: []byte(`[
			{"id":"layout.title","translation":"Blog"}
		]`)},
		"messages/seo.en.json": &fstest.MapFile{Data: []byte(`[
			{
				"id":"seo.author.description",
				"translation":"Browse notes by {{.Author}}.",
				"args":[{"name":"Author","type":"string"}]
			}
		]`)},
		"messages/layout.de.json": &fstest.MapFile{Data: []byte(`[
			{"id":"layout.title","translation":"Blog DE"}
		]`)},
		"messages/seo.de.json": &fstest.MapFile{Data: []byte(`[
			{"id":"seo.author.description","translation":"Notizen von {{.Author}}."}
		]`)},
	}

	definitions, canonicalLocale, err := LoadMessageDefinitions(
		filesystem,
		[]string{"messages/layout.en.json", "messages/seo.en.json", "messages/layout.de.json", "messages/seo.de.json"},
		"en",
	)
	require.NoError(t, err)
	require.Equal(t, "en", canonicalLocale)
	require.Len(t, definitions["en"], 2)
	require.Len(t, definitions["de"], 2)
	require.Equal(t, "layout.title", definitions["en"][0].ID)
	require.Equal(t, "seo.author.description", definitions["en"][1].ID)
	require.Len(t, definitions["en"][1].Args, 1)
}

func TestValidateMessageKeyParityPasses(t *testing.T) {
	t.Parallel()

	payload := buildLocalePayload(t, []string{"one", "two"})
	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: payload},
		"messages/active.de.json": &fstest.MapFile{Data: payload},
	}

	err := ValidateMessageKeyParity(filesystem, []string{
		"messages/active.en.json",
		"messages/active.de.json",
	}, []string{"one", "two"})
	require.NoError(t, err)
}

func TestValidateMessageKeyParityRejectsMissingKey(t *testing.T) {
	t.Parallel()

	payload := buildLocalePayload(t, []string{"one"})
	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: payload},
	}

	err := ValidateMessageKeyParity(filesystem, []string{"messages/active.en.json"}, []string{"one", "two"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing=")
}

func TestValidateMessageKeyParityRejectsExtraKey(t *testing.T) {
	t.Parallel()

	payload := buildLocalePayload(t, []string{"one", "two", "extra"})
	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: payload},
	}

	err := ValidateMessageKeyParity(filesystem, []string{"messages/active.en.json"}, []string{"one", "two"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "extra=")
}

func TestValidateMessageKeyParityRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	entries := []localeEntry{
		{ID: "one", Translation: "first"},
		{ID: "one", Translation: "second"},
	}
	payload, err := json.Marshal(entries)
	require.NoError(t, err)
	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: payload},
	}

	err = ValidateMessageKeyParity(filesystem, []string{"messages/active.en.json"}, []string{"one"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate message id")
}

func TestValidateMessageCatalogRejectsDuplicateIDsAcrossShards(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/layout.en.json": &fstest.MapFile{Data: []byte(`[
			{"id":"layout.title","translation":"Blog"}
		]`)},
		"messages/seo.en.json": &fstest.MapFile{Data: []byte(`[
			{"id":"layout.title","translation":"Duplicate"}
		]`)},
	}

	err := ValidateMessageCatalog(
		filesystem,
		[]string{"messages/layout.en.json", "messages/seo.en.json"},
		"en",
		[]string{"layout.title"},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate message id")
}

func TestParseCanonicalMessagesRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	_, err := ParseCanonicalMessages([]byte(`[
		{"id":"a.b","translation":"x"},
		{"id":"a.b","translation":"y"}
	]`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate message id")
}

func TestParseCanonicalMessagesAcceptsTypedArgs(t *testing.T) {
	t.Parallel()

	source := []byte(`[
		{
			"id":"seo.author.description",
			"translation":"Browse notes by {{.Author}}.",
			"args":[{"name":"Author","type":"string"}]
		}
	]`)

	messages, err := ParseCanonicalMessages(source)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Len(t, messages[0].Args, 1)
	require.Equal(t, "Author", messages[0].Args[0].Name)
}

func TestParseCanonicalMessagesRejectsUndeclaredPlaceholder(t *testing.T) {
	t.Parallel()

	_, err := ParseCanonicalMessages([]byte(`[
		{"id":"seo.author.description","translation":"Browse notes by {{.Author}}."}
	]`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "undeclared placeholder")
}

func TestParseCanonicalMessagesRejectsUnsupportedTemplateAction(t *testing.T) {
	t.Parallel()

	source := []byte(`[
		{
			"id":"seo.author.description",
			"translation":"Browse notes by {{printf \"%s\" .Author}}.",
			"args":[{"name":"Author","type":"string"}]
		}
	]`)

	_, err := ParseCanonicalMessages(source)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported template action")
}

func TestValidateMessageCatalogRejectsPlaceholderParityMismatch(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/active.en.json": &fstest.MapFile{Data: []byte(`[
			{
				"id":"seo.author.description",
				"translation":"Browse notes by {{.Author}}.",
				"args":[{"name":"Author","type":"string"}]
			}
		]`)},
		"messages/active.de.json": &fstest.MapFile{Data: []byte(`[
			{"id":"seo.author.description","translation":"Browse notes by {{.Autor}}."}
		]`)},
	}

	err := ValidateMessageCatalog(
		filesystem,
		[]string{"messages/active.en.json", "messages/active.de.json"},
		"en",
		[]string{"seo.author.description"},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "placeholder mismatch")
}

func buildLocalePayload(t *testing.T, keys []string) []byte {
	t.Helper()

	entries := make([]localeEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, localeEntry{
			ID:          key,
			Translation: key,
		})
	}

	payload, err := json.Marshal(entries)
	require.NoError(t, err)
	return payload
}
