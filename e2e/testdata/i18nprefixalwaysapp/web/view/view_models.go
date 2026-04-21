package runtime

import (
	i18n "example.com/no-js-e2e/i18nprefixalwaysapp/web/generated/i18n"
	frameworki18n "github.com/RevoTale/no-js/framework/i18n"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type HomePageView struct {
	RootLayoutView
	Heading         string
	Kicker          string
	Locale          string
	SwitchToEnglish string
	SwitchToGerman  string
}

type GreetPageView struct {
	RootLayoutView
	Heading         string
	Description     string
	Locale          string
	SwitchToEnglish string
	SwitchToGerman  string
}

func NewNotFoundView(messages frameworki18n.Context[i18n.Key]) RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView(messages frameworki18n.Context[i18n.Key]) RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return nil
}
