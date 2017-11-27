package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/varddum/syndication/models"
	"github.com/varddum/syndication/plugins"
)

func helloWorldHandler(c plugins.APICtx, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
}

func entriesHandler(c plugins.APICtx, w http.ResponseWriter, r *http.Request) {
	if c.HasUser() {
		entries, err := c.User.Entries(true, models.Any)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(entries) == 0 {
			fmt.Fprintf(w, "Nothing new!\n")
			return
		}

		js, err := json.Marshal(entries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func Initialize() (plugins.Plugin, error) {
	plugin := plugins.NewAPIPlugin("API Example")

	plugin.RegisterEndpoint(plugins.Endpoint{
		NeedsUser: false,
		Path:      "/hello_world",
		Method:    "GET",
		Group:     "api_test",
		Handler:   helloWorldHandler,
	})

	plugin.RegisterEndpoint(plugins.Endpoint{
		NeedsUser: true,
		Path:      "/entries",
		Method:    "GET",
		Group:     "api_test",
		Handler:   entriesHandler,
	})

	return plugin, nil
}

func Shutdown() {

}
