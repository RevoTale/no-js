package resolvers

import (
	"context"
	"net/http"

	"example.com/no-js-e2e/clientassetsslotgroupapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/metagen"
)

func (Resolver) MetaGenGroupLabGroupExperimentsOpsLayout(
	meta framework.MetaContext[*view.Context],
	params GroupLabGroupExperimentsOpsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Grouped Ops Layout"}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsParams,
) (view.OpsLayoutView, error) {
	return view.OpsLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops"},
	}, nil
}

func (Resolver) MetaGenGroupLabGroupExperimentsOpsPage(
	meta framework.MetaContext[*view.Context],
	params GroupLabGroupExperimentsOpsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Grouped Ops"}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsParams,
) (view.OpsPageView, error) {
	return view.OpsPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops"},
		Heading:        "Grouped Ops",
	}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsSlotPanelLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsSlotPanelParams,
) (view.OpsPanelLayoutView, error) {
	return view.OpsPanelLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops Panel"},
	}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsSlotPanelDefault(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsSlotPanelParams,
) (view.OpsPanelDefaultView, error) {
	return view.OpsPanelDefaultView{Label: "Ops panel fallback"}, nil
}

func (Resolver) MetaGenGroupLabGroupExperimentsOpsReportsLayout(
	meta framework.MetaContext[*view.Context],
	params GroupLabGroupExperimentsOpsReportsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Grouped Ops Reports Layout"}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsReportsLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsReportsParams,
) (view.OpsReportsLayoutView, error) {
	return view.OpsReportsLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops Reports"},
	}, nil
}

func (Resolver) MetaGenGroupLabGroupExperimentsOpsReportsPage(
	meta framework.MetaContext[*view.Context],
	params GroupLabGroupExperimentsOpsReportsParams,
) (metagen.Metadata, error) {
	return metagen.Metadata{Title: "Grouped Ops Reports"}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsReportsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsReportsParams,
) (view.OpsReportsPageView, error) {
	return view.OpsReportsPageView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops Reports"},
		Heading:        "Grouped Ops Reports",
	}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsSlotPanelGroupParallelReportsLayout(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsSlotPanelGroupParallelReportsParams,
) (view.OpsPanelReportsLayoutView, error) {
	return view.OpsPanelReportsLayoutView{
		RootLayoutView: view.RootLayoutView{PageTitle: "Grouped Ops Panel Reports"},
	}, nil
}

func (Resolver) ResolveGroupLabGroupExperimentsOpsSlotPanelGroupParallelReportsPage(
	ctx context.Context,
	appCtx *view.Context,
	r *http.Request,
	params GroupLabGroupExperimentsOpsSlotPanelGroupParallelReportsParams,
) (view.OpsPanelReportsPageView, error) {
	return view.OpsPanelReportsPageView{Label: "Ops reports panel"}, nil
}
