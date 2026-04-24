package view

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type RootPageView struct {
	RootLayoutView
	Heading string
}

func NewNotFoundView() RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView() RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}
