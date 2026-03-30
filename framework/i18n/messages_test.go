package i18n

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

type localeEntry struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

func TestDiscoverMessageFilesSorted(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"messages/active.de.json": &fstest.MapFile{Data: []byte("[]")},
		"messages/active.en.json": &fstest.MapFile{Data: []byte("[]")},
	}

	files, err := DiscoverMessageFiles(filesystem)
	require.NoError(t, err)
	require.Equal(t, "messages/active.de.json,messages/active.en.json", strings.Join(files, ","))
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

func TestParseCanonicalMessagesRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	_, err := ParseCanonicalMessages([]byte(`[
		{"id":"a.b","translation":"x"},
		{"id":"a.b","translation":"y"}
	]`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate message id")
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
