package routes

import (
	"net/http"

	runtime "example.com/templcssapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Robots(runtimeCtx framework.RuntimeContext[*runtime.Context], r *http.Request) (discovery.Robots, error) {
	return discovery.Robots{}, nil
}
