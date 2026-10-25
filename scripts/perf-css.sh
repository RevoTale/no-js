#!/usr/bin/env bash
set -euo pipefail
probe_dir=$(mktemp -d)
trap 'rm -rf "$probe_dir"' EXIT
templ_version=$(go list -m -f '{{.Version}}' github.com/a-h/templ)
cat > "$probe_dir/styles.templ" <<'TEMPL'
package cssprobe
var color = "red"
css Static() {
 color: red;
 padding: 1rem;
}
css Dynamic() {
 color: { color };
}
TEMPL
cat > "$probe_dir/go.mod" <<MOD
module cssprobe

go 1.25.0

require github.com/a-h/templ $templ_version
MOD
cat > "$probe_dir/styles_test.go" <<'GO'
package cssprobe
import (
 "testing"
 "github.com/a-h/templ"
)
var sink templ.CSSClass
func BenchmarkStaticCSS(b *testing.B) {
 b.ReportAllocs()
 for b.Loop() { sink = Static() }
}
func TestCSSContract(t *testing.T) {
 color = "red"
 expectedID := templ.CSSID("Static", "color:red;padding:1rem;")
 actual := Static().(templ.ComponentCSSClass)
 if actual.ID != expectedID || string(actual.Class) != "."+expectedID+"{color:red;padding:1rem;}" {
  t.Fatalf("unexpected CSS: %#v", actual)
 }
 before := Dynamic().(templ.ComponentCSSClass)
 color = "blue"
 after := Dynamic().(templ.ComponentCSSClass)
 if before.ID == after.ID { t.Fatal("dynamic CSS was frozen") }
}
GO
# Compare the same fixture, toolchain, and pinned templ dependency.
(cd "$probe_dir" && go run "github.com/a-h/templ/cmd/templ@$templ_version" generate -path .)
(cd "$probe_dir" && go test -mod=mod -bench BenchmarkStaticCSS -benchmem -count 3)
go run ./cmd/templgen -base "$probe_dir" -path "$probe_dir"
(cd "$probe_dir" && go test -mod=mod -bench BenchmarkStaticCSS -benchmem -count 3)
