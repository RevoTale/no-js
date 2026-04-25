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

func requireNoOriginalClassSelector(t *testing.T, css string, className string) {
	t.Helper()

	pattern := regexp.MustCompile(`(^|[,{>+~\s])\.` + regexp.QuoteMeta(className) + `([,{>+~:#.\[]|$)`)
	require.NotRegexp(t, pattern, css)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func requireContainsInlineClassRule(t *testing.T, html string, className string, token string) {
	t.Helper()

	pattern := `\.` + regexp.QuoteMeta(className) +
		`\{[^}]*` + regexp.QuoteMeta(token) + `[^}]*\}`
	rulePattern := regexp.MustCompile(pattern)
	require.Regexp(t, rulePattern, html)
}
