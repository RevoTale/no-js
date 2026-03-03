package metagen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const htmxAfterSettleHeader = "HX-Trigger-After-Settle"

func BuildHTMXPatch(meta Metadata) (Patch, error) {
	normalized := Normalize(meta)
	title, headHTML, err := renderManagedHead(normalized, false)
	if err != nil {
		return Patch{}, err
	}

	return Patch{
		Title: title,
		Head:  headHTML,
	}, nil
}

func WriteHTMXHeaders(w http.ResponseWriter, patch Patch) error {
	if w == nil {
		return nil
	}

	payload := map[string]Patch{
		HTMXPatchEvent: patch,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal htmx metadata patch: %w", err)
	}

	header := w.Header()
	existing := strings.TrimSpace(header.Get(htmxAfterSettleHeader))
	if existing == "" {
		header.Set(htmxAfterSettleHeader, string(encoded))
		return nil
	}

	merged, mergeErr := mergeJSONHeader(existing, payload)
	if mergeErr != nil {
		header.Set(htmxAfterSettleHeader, string(encoded))
		return nil
	}
	header.Set(htmxAfterSettleHeader, merged)
	return nil
}

func mergeJSONHeader(existing string, newPayload map[string]Patch) (string, error) {
	current := make(map[string]json.RawMessage)
	if err := json.Unmarshal([]byte(existing), &current); err != nil {
		return "", err
	}

	for eventName, payload := range newPayload {
		encodedPayload, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		current[eventName] = encodedPayload
	}

	encoded, err := json.Marshal(current)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
