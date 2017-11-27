package main

import (
	"fmt"
	"html"
	"net/http"

	"github.com/varddum/syndication/plugins"
)

func subscriptionsHandler(c plugins.UserCtx, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func Initialize() (plugins.Plugin, error) {
	plugin := plugins.NewAPIPlugin("API Example")
	plugin.RegisterEndpoint(plugins.Endpoint{
		Path:    "/fever/subscriptions",
		Method:  "GET",
		Group:   "api_test",
		Handler: subscriptionsHandler,
	})

	return plugin, nil
}

func Shutdown() {

}
