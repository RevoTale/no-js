// Package templruntime supports Go templates emitted by the no-js generator.
package templruntime

import (
	"strings"

	"github.com/a-h/templ"
)

// ClassNames matches templ.Classes(classes...).String(). Complex inputs retain
// upstream behavior; generated templates use this helper to avoid common allocations.
func ClassNames(classes ...any) string {
	// Keep the scan bounded; large lists use upstream linear map processing.
	var names [8]string
	if len(classes) > len(names) {
		return templ.Classes(classes...).String()
	}
	count, capacity := 0, 0
	for _, class := range classes {
		name, simple := simpleClassName(class)
		if !simple {
			return templ.Classes(classes...).String()
		}
		duplicate := false
		for _, previous := range names[:count] {
			if previous == name {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		names[count] = name
		count++
		capacity += len(name)
	}
	if count == 0 {
		return ""
	}
	if count == 1 {
		return names[0]
	}
	var builder strings.Builder
	builder.Grow(capacity + count - 1)
	for index, name := range names[:count] {
		if index > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(name)
	}
	return builder.String()
}

func simpleClassName(class any) (string, bool) {
	switch value := class.(type) {
	case string:
		return value, true
	case templ.ConstantCSSClass:
		return value.ClassName(), true
	case templ.ComponentCSSClass:
		return value.ClassName(), true
	default:
		return "", false
	}
}
