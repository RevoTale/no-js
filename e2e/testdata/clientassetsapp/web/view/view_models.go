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

type ComplexPageView struct {
	RootLayoutView
	Heading string
}

type SectionSummaryPageView struct {
	RootLayoutView
	Heading string
}

type SectionAdminLayoutView struct {
	RootLayoutView
}

type SectionAdminPageView struct {
	RootLayoutView
	Heading string
}
