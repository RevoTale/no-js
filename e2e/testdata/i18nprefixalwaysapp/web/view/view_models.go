package view

type RootLayoutView struct {
	PageTitle         string
	NotFoundHeading   string
	ErrorHeading      string
	SystemLocale      string
	SystemRequestPath string
	NotFoundSource    string
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

type FailPageView struct {
	RootLayoutView
}

func NewNotFoundView() RootLayoutView {
	return RootLayoutView{
		PageTitle:       "Not Found",
		NotFoundHeading: "Not Found",
	}
}

func NewErrorView() RootLayoutView {
	return RootLayoutView{
		PageTitle:    "Error",
		ErrorHeading: "Error",
	}
}
