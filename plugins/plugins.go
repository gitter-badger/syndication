/*
  Copyright (C) 2017 Jorge Martinez Hernandez

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.
  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package plugins

import (
	"net/http"
	"plugin"

	log "github.com/sirupsen/logrus"
)

// RequestHandler represents the function type for Endpoint Handlers in API Plugins
type RequestHandler = func(APICtx, http.ResponseWriter, *http.Request)

type InitFunc = func() (Plugin, error)

type (
	// Plugin collects properties for all plugins
	Plugin interface {
		Path() string
		Name() string
	}

	// Endpoint represents an API Endpoint that can be registered by an API Plugin.
	Endpoint struct {
		NeedsUser bool
		Path      string
		Method    string
		Group     string
		Handler   RequestHandler
	}

	// APIPlugin collects information on an API Plugin and the endpoints it registers.
	APIPlugin struct {
		name      string
		endpoints []Endpoint
		path      string
	}

	// Plugins manages the available plugins configured and registered for a Syndication instance.
	Plugins struct {
		apiPlugins []APIPlugin
	}

	APIPluginError struct {
		ErrorMsg string
	}
)

func (e APIPluginError) Error() string {
	return e.ErrorMsg
}

func (p APIPlugin) Path() string {
	return p.path
}

func (p APIPlugin) Name() string {
	return p.name
}

func NewAPIPlugin(name string) APIPlugin {
	return APIPlugin{name: name}
}

func (p *APIPlugin) Endpoints() []Endpoint {
	return p.endpoints
}

func (p *APIPlugin) RegisterEndpoint(endpnt Endpoint) error {
	if endpnt.Handler == nil {
		return APIPluginError{"A handler is required."}
	}

	if endpnt.Method == "" {
		return APIPluginError{"A method is required."}
	}

	if endpnt.Path == "" {
		return APIPluginError{"A path is required."}
	}

	if p.checkConflictingPaths(endpnt) {
		return APIPluginError{"The path " + endpnt.Path + "for method " + endpnt.Method + " already exists."}
	}

	p.endpoints = append(p.endpoints, endpnt)

	return nil
}

func (p APIPlugin) checkConflictingPaths(incomingEndpnt Endpoint) bool {
	// TODO: This will be a linear search for now.
	for _, endpnt := range p.endpoints {
		if endpnt.Path == incomingEndpnt.Path && endpnt.Method == incomingEndpnt.Method {
			return true
		}
	}

	return false
}

func NewPlugins(pluginPaths []string) Plugins {
	plugins := Plugins{}

	plugins.loadPlugins(pluginPaths)

	return plugins
}

func (s *Plugins) loadPlugins(paths []string) {
	for _, path := range paths {
		plgn, err := plugin.Open(path)
		if err != nil {
			log.Error(err, ". Skipping.")
			continue
		}

		initFuncSymb, err := plgn.Lookup("Initialize")
		if err != nil {
			log.Error(err, ". Skipping.")
			continue
		}

		initFunc, ok := initFuncSymb.(InitFunc)
		if !ok {
			log.Error("Invalid Initialization function.")
			continue
		}

		incomingPlgn, err := initFunc()
		if err != nil {
			log.Error(err, ". Skpping.")
			continue
		}

		switch t := incomingPlgn.(type) {
		case APIPlugin:
			s.apiPlugins = append(s.apiPlugins, t)
		default:
			log.Error("Unrecognized plugin type.")
		}

	}
}

func (s *Plugins) APIPlugins() []APIPlugin {
	return s.apiPlugins
}
