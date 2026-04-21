package e2e

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustMatchString(t *testing.T, pattern string, value string) string {
	t.Helper()

	re := regexp.MustCompile(pattern)
	match := re.FindString(value)
	require.NotEmpty(t, match, "pattern %q not found", pattern)
	return match
}

func classesForOpeningTag(t *testing.T, tag string) []string {
	t.Helper()

	re := regexp.MustCompile(`class="([^"]+)"`)
	matches := re.FindStringSubmatch(tag)
	require.Len(t, matches, 2, "class attribute missing from %s", tag)
	return strings.Fields(matches[1])
}

func requireClassWithToken(t *testing.T, css string, classes []string, token string) string {
	t.Helper()

	for _, className := range classes {
		pattern := `\.` + regexp.QuoteMeta(className) +
			`\{[^}]*` + regexp.QuoteMeta(token) + `[^}]*\}`
		rulePattern := regexp.MustCompile(pattern)
		if rulePattern.MatchString(css) {
			return className
		}
	}

	t.Fatalf("no css rule for classes %v contains %q", classes, token)
	return ""
}

func requireContainsInlineClassRule(t *testing.T, html string, className string, token string) {
	t.Helper()

	pattern := `\.` + regexp.QuoteMeta(className) +
		`\{[^}]*` + regexp.QuoteMeta(token) + `[^}]*\}`
	rulePattern := regexp.MustCompile(pattern)
	require.Regexp(t, rulePattern, html)
}
