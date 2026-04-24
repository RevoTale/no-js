package view

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"example.com/no-js-e2e/templrulesapp/web/components"
	"github.com/a-h/templ"
)

type RootLayoutView struct {
	PageTitle string
}

func (view RootLayoutView) LayoutPageTitle() string {
	return view.PageTitle
}

type CardPageView struct {
	RootLayoutView
	Title  string
	Urgent bool
}

type BoardPageView struct {
	RootLayoutView
	Title string
	Body  string
}

type MeterPageView struct {
	RootLayoutView
	Progress string
	Accent   string
}

type ProgressPageView struct {
	RootLayoutView
	Percent int
}

type DashboardPageView struct {
	RootLayoutView
	Title   string
	Percent int
}

type StreamPageView struct {
	RootLayoutView
	Stream *StreamState
}

func NewNotFoundView() RootLayoutView {
	return RootLayoutView{PageTitle: "Not Found"}
}

func NewErrorView() RootLayoutView {
	return RootLayoutView{PageTitle: "Error"}
}

func TemplCSSVariants() []templ.CSSClass {
	return []templ.CSSClass{
		components.ProgressBar(72),
	}
}

func StreamBody(state *StreamState) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "first"); err != nil {
			return err
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("writer does not implement http.Flusher")
		}
		flusher.Flush()

		if state != nil {
			state.Wait()
		}

		_, err := io.WriteString(w, "second")
		return err
	})
}
