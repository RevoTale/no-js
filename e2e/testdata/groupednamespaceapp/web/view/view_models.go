package runtime

import (
	"example.com/no-js-e2e/groupednamespaceapp/web/components"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type DiscoverPageView struct {
	RootLayoutView
	Label    string
	Title    string
	Progress int
}

func NewNotFoundView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		components.ProgressBar(64),
		components.ProgressBar(80),
	}
}
