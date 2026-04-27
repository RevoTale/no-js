package routes

import (
	"net/http"

	"example.com/no-js-e2e/optionalcatchallapp/web/view"
	"github.com/RevoTale/no-js/framework"
	"github.com/RevoTale/no-js/framework/discovery"
)

func Robots(runtimeCtx framework.RuntimeContext[*view.Context], r *http.Request) (discovery.Robots, error) {
	return discovery.Robots{}, nil
}
