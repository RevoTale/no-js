package approutegen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func safeIdentifier(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "value"
	}

	normalized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, trimmed)
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		return "value"
	}
	return strings.ToLower(normalized)
}

func pascalToken(value string) string {
	normalized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return ' '
	}, strings.TrimSpace(value))

	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return ""
	}

	builder := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(strings.ToLower(part[1:]))
		}
	}

	return builder.String()
}

func writef(buffer *bytes.Buffer, pattern string, args ...any) {
	_, _ = fmt.Fprintf(buffer, pattern, args...)
}

func dedupeSorted(values []string) []string {
	sort.Strings(values)
	if len(values) == 0 {
		return values
	}

	out := make([]string, 0, len(values))
	previous := ""
	for idx, value := range values {
		if idx > 0 && value == previous {
			continue
		}
		out = append(out, value)
		previous = value
	}

	return out
}
