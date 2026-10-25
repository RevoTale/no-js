package templgen

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/parser/v2"
)

// precomputeStaticCSS preserves templ signatures and class IDs while moving
// constant CSS construction out of the request path. Expression properties stay
// with the upstream generator, even when their CSS function has no parameters.
func precomputeStaticCSS(file *parser.TemplateFile) {
	for index, node := range file.Nodes {
		css, ok := node.(*parser.CSSTemplate)
		if !ok {
			continue
		}
		// Receiver methods can share a name across types and files. Keep their
		// upstream generation rather than creating conflicting package variables.
		if strings.HasPrefix(strings.TrimSpace(css.Expression.Value), "(") {
			continue
		}
		value, ok := constantCSS(css)
		if !ok {
			continue
		}
		id := templ.CSSID(css.Name, value)
		variable := "templ_nojs_static_css_" + css.Name
		code := fmt.Sprintf(`var %s templ.CSSClass = templ.ComponentCSSClass{
ID: %q,
Class: templ.SafeCSS(%q),
}
func %s templ.CSSClass { return %s }
`, variable, id, "."+id+"{"+value+"}", css.Expression.Value, variable)
		file.Nodes[index] = &parser.TemplateFileGoExpression{
			Expression: parser.Expression{Value: code, Range: css.Range},
		}
	}
}

func constantCSS(css *parser.CSSTemplate) (string, bool) {
	var value strings.Builder
	for _, property := range css.Properties {
		constant, ok := property.(*parser.ConstantCSSProperty)
		if !ok {
			return "", false
		}
		value.WriteString(constant.String(true))
	}
	return value.String(), true
}
