package templgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

func TestRunGeneratesOutputFromPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourcePath := filepath.Join(root, "hello.templ")
	writeTestFile(t, sourcePath, "package sample\n\ntempl Hello() { <div>hello</div> }\n")

	err := Run(Config{
		Paths:    []string{root},
		BasePath: root,
	})
	require.NoError(t, err)

	generatedPath := filepath.Join(root, "hello_templ.go")
	_, statErr := os.Stat(generatedPath)
	require.NoError(t, statErr)
}

func TestRunReturnsErrorForEmptySelection(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := Run(Config{
		BasePath: root,
	})
	require.Error(t, err)
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))
}

func TestRunPrecomputesOnlyConstantCSS(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourcePath := filepath.Join(root, "styles.templ")
	source := `package sample
var color = "red"
css Static() {
 color: red;
 padding: 1rem;
}
css Dynamic() {
 color: { color };
}
css Parameterized(color string) {
 color: { color };
}
`
	writeTestFile(t, sourcePath, source)
	require.NoError(t, Run(Config{Files: []string{sourcePath}, BasePath: root}))
	generated, err := os.ReadFile(filepath.Join(root, "styles_templ.go"))
	require.NoError(t, err)
	output := string(generated)
	require.Contains(t, output, "var templ_nojs_static_css_Static templ.CSSClass")
	require.Contains(t, output, "return templ_nojs_static_css_Static")
	require.Contains(t, output, `ID:    "`+templ.CSSID("Static", "color:red;padding:1rem;")+`"`)
	require.Contains(t, output, `templ.SanitizeCSS(`)
	require.Contains(t, output, "func Parameterized(color string) templ.CSSClass")
	require.NotContains(t, output, "templ_nojs_static_css_Dynamic")
	require.NotContains(t, output, "templ_nojs_static_css_Parameterized")
}

func TestRunOptimizesClassLists(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourcePath := filepath.Join(root, "classes.templ")
	writeTestFile(t, sourcePath, `package sample
var templ_nojs_runtime = "hook"
templ Example() { <div class={ templ_nojs_runtime, "body", templ.KV("active", true) }></div> }
`)
	require.NoError(t, Run(Config{Files: []string{sourcePath}, BasePath: root}))
	generated, err := os.ReadFile(filepath.Join(root, "classes_templ.go"))
	require.NoError(t, err)
	output := string(generated)
	require.Contains(t, output, `"github.com/RevoTale/no-js/v2/framework/templruntime"`)
	require.Contains(t, output, "templ_nojs_runtime_.ClassNames(")
	require.NotContains(t, output, "templ.Classes(")
	require.Contains(t, output, "templ.EscapeString(")
}

func TestClassOptimizationPreservesUserGoFunctions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourcePath := filepath.Join(root, "manual.templ")
	writeTestFile(t, sourcePath, `package sample
type customFactory struct{}
type customClassList struct{}
func (customFactory) Classes() customClassList { return customClassList{} }
func (customClassList) String() string { return "manual" }
func Manual() string {
 templ := customFactory{}
 return templ.Classes().String()
}
templ Example() { <div class={ "hook", "body" }></div> }
`)
	require.NoError(t, Run(Config{Files: []string{sourcePath}, BasePath: root}))
	generated, err := os.ReadFile(filepath.Join(root, "manual_templ.go"))
	require.NoError(t, err)
	require.Contains(t, string(generated), "return templ.Classes().String()")
	require.Contains(t, string(generated), ".ClassNames(")
}

func TestClassOptimizationPreservesShadowedTemplInsideTemplate(t *testing.T) {
	t.Parallel()
	source := []byte(`package sample
import "github.com/a-h/templ"
import "github.com/a-h/templ/runtime"
func Example() templ.Component {
 return templruntime.GeneratedTemplate(func(input templruntime.GeneratedComponentInput) error {
  _ = func() string {
   templ := customFactory{}
   return templ.Classes().String()
  }()
  _ = templ.Classes("hook", "body").String()
  return nil
 })
}
`)
	output, err := optimizeClassNames(source)
	require.NoError(t, err)
	require.Contains(t, string(output), "return templ.Classes().String()")
	require.Contains(t, string(output), `.ClassNames("hook", "body")`)
}

func TestRunPreservesCSSMethodsWithSameName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourcePath := filepath.Join(root, "methods.templ")
	writeTestFile(t, sourcePath, `package sample
type A[T any] struct{}
type B[T any] struct{}
css (a A[T]) Style() {
 color: red;
}
css (b *B[T]) Style() {
 color: blue;
}
`)
	require.NoError(t, Run(Config{Files: []string{sourcePath}, BasePath: root}))
	generated, err := os.ReadFile(filepath.Join(root, "methods_templ.go"))
	require.NoError(t, err)
	output := string(generated)
	require.Contains(t, output, "func (a A[T]) Style() templ.CSSClass")
	require.Contains(t, output, "func (b *B[T]) Style() templ.CSSClass")
	require.NotContains(t, output, "templ_nojs_static_css_Style")
}
