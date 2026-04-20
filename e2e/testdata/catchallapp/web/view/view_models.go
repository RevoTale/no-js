package runtime

import "github.com/a-h/templ"

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type CatchAllPageView struct {
	RootLayoutView
	Joined string
	Depth  string
}

func NewNotFoundView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView(messages *Messages) RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return nil
}
