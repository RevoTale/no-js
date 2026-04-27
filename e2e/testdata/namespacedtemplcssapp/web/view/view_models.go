package view

import (
	"example.com/no-js-e2e/namespacedtemplcssapp/web/components/statchip"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		statchip.ProgressBar(72),
	}
}
