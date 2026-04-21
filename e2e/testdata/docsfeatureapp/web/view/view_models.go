package runtime

import (
	"example.com/no-js-e2e/docsfeatureapp/web/components"
	i18n "example.com/no-js-e2e/docsfeatureapp/web/generated/i18n"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
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

func NewNotFoundView(messages frameworki18n.Context[i18n.Key]) RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView(messages frameworki18n.Context[i18n.Key]) RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		components.ProgressBar(64),
		components.ProgressBar(72),
	}
}
