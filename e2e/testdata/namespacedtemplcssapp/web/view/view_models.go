package view

import (
	"example.com/no-js-e2e/namespacedtemplcssapp/web/components"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

func NewNotFoundView() RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView() RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		components.ProgressBar(72),
	}
}
