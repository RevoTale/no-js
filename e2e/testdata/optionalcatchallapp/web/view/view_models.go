package view

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type OptionalCatchAllPageView struct {
	RootLayoutView
	Joined string
	Depth  string
}
