package view

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type OpsLayoutView struct {
	RootLayoutView
}

type OpsPageView struct {
	RootLayoutView
	Heading string
}

type OpsPanelDefaultView struct {
	Label string
}

type OpsPanelLayoutView struct {
	RootLayoutView
}

type OpsReportsLayoutView struct {
	RootLayoutView
}

type OpsReportsPageView struct {
	RootLayoutView
	Heading string
}

type OpsPanelReportsLayoutView struct {
	RootLayoutView
}

type OpsPanelReportsPageView struct {
	Label string
}
