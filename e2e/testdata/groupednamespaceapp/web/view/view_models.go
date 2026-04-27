package view

import (
	"example.com/no-js-e2e/groupednamespaceapp/web/components/discovercard"
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

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		discovercard.ProgressBar(64),
		discovercard.ProgressBar(80),
	}
}
