package view

import (
	"example.com/no-js-e2e/docsfeatureapp/web/components/profile"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type AuthorPageView struct {
	RootLayoutView
	Heading         string
	Description     string
	SharedA         string
	SharedB         string
	SwitchToEnglish string
	LoadCount       string
	Progress        int
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		profile.ProgressBar(64),
		profile.ProgressBar(72),
	}
}
