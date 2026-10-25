package templruntime

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

var classNamesSink string

func TestClassNamesMatchesTempl(t *testing.T) {
	t.Parallel()
	component := templ.ComponentCSSClass{ID: "card_123", Class: ".card_123{color:red;}"}
	constant := templ.ConstantCSSClass("button")
	cases := [][]any{
		nil, {"a"}, {"a", "b", "a"}, {"", "a", ""}, {"a b", "a", "a b"},
		{component, constant, "hook", component}, {"a", templ.KV("a", false)},
		{map[string]bool{"z": true, "a": true}}, {[]string{"a", "b"}},
		{templ.CSSClasses{"a", component}}, {templ.KV[templ.CSSClass, bool](component, true)},
		{nil}, {struct{}{}}, {[]templ.CSSClass{component, constant}},
	}
	for _, classes := range cases {
		require.Equal(t, templ.Classes(classes...).String(), ClassNames(classes...))
	}
}

func TestClassNamesDoesNotEvaluateFunctionsTwice(t *testing.T) {
	calls := 0
	class := func() templ.CSSClass { calls++; return templ.ConstantCSSClass("dynamic") }
	require.Equal(t, "before dynamic", ClassNames("before", class))
	require.Equal(t, 1, calls)
}

func BenchmarkClassNames(b *testing.B) {
	classes := []any{"hook", templ.ConstantCSSClass("body"), templ.ComponentCSSClass{ID: "card_123"}, "hook"}
	b.ReportAllocs()
	for b.Loop() {
		classNamesSink = ClassNames(classes...)
	}
}

func TestClassNamesAllocationBudget(t *testing.T) {
	classes := []any{"hook", templ.ConstantCSSClass("body"), templ.ComponentCSSClass{ID: "card_123"}, "hook"}
	allocations := testing.AllocsPerRun(100, func() { classNamesSink = ClassNames(classes...) })
	require.Equal(t, "hook body card_123", classNamesSink)
	require.LessOrEqual(t, allocations, float64(1))
}

func TestClassNamesLargeList(t *testing.T) {
	classes := make([]any, 128)
	for index := range classes {
		classes[index] = "same"
	}
	require.Equal(t, templ.Classes(classes...).String(), ClassNames(classes...))
}

func TestDuplicateClassNamesDoNotCopyLongStrings(t *testing.T) {
	name := strings.Repeat("duplicate", 4096)
	classes := []any{name, name, name, name, name, name, name, name}
	allocations := testing.AllocsPerRun(100, func() { classNamesSink = ClassNames(classes...) })
	require.Equal(t, name, classNamesSink)
	require.Zero(t, allocations)
}
