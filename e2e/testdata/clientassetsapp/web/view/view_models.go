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

type AboutPageView struct {
	RootLayoutView
	Heading string
}

type SectionLayoutView struct {
	RootLayoutView
}

type SectionPageView struct {
	RootLayoutView
	Heading string
}
