package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newPerformanceRuntime(tb testing.TB) *Runtime[string] {
	tb.Helper()
	runtime, err := NewRuntime[string](Config{
		Locales: []string{"en", "uk", "de"}, DefaultLocale: "en", PrefixMode: PrefixAsNeeded,
		DisplayLabels: map[string]string{"en": "English", "uk": "Українська", "de": "Deutsch"},
		DisplayOrder:  []string{"uk", "en", "de"},
	}, nil, nil)
	require.NoError(tb, err)
	return runtime
}

func BenchmarkRuntimePathFor(b *testing.B) {
	ctx := newPerformanceRuntime(b).Context(nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = ctx.PathFor("uk", "/notes/example")
	}
}

func TestRuntimePathForAllocationBudget(t *testing.T) {
	ctx := newPerformanceRuntime(t).Context(nil, nil)
	var got string
	allocations := testing.AllocsPerRun(100, func() {
		got = ctx.PathFor("uk", "/notes/example")
	})
	require.Equal(t, "/uk/notes/example", got)
	require.LessOrEqual(t, allocations, float64(1), "config must be normalized at construction")
}

func TestRuntimeConfigSnapshot(t *testing.T) {
	runtime := newPerformanceRuntime(t)
	cfg := runtime.Config()
	cfg.Locales[1] = "xx"
	cfg.DisplayOrder[0] = "xx"
	cfg.DisplayLabels["uk"] = "changed"
	ctx := runtime.Context(nil, nil)
	require.Equal(t, "/uk/notes/example", ctx.PathFor("uk", "/notes/example"))
	require.Equal(t, "Українська", runtime.Config().DisplayLabels["uk"])
	require.Equal(t, "uk", runtime.Config().DisplayOrder[0])
}

func BenchmarkLiteralMessage(b *testing.B) {
	message := CompiledMessage{Parts: []CompiledMessagePart{{Text: "Software systems, tools and open source."}}}
	b.ReportAllocs()
	for b.Loop() {
		_ = message.Render(nil)
	}
}

func TestLiteralMessageDoesNotAllocate(t *testing.T) {
	message := CompiledMessage{Parts: []CompiledMessagePart{{Text: "Software systems, tools and open source."}}}
	var rendered string
	allocations := testing.AllocsPerRun(100, func() { rendered = message.Render(nil) })
	require.Equal(t, message.Parts[0].Text, rendered)
	require.Zero(t, allocations)
}
