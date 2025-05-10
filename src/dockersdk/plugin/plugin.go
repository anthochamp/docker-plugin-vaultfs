// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"net/http"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type PluginHandlerFn func(util.HttpRequest) error

type Plugin struct {
	serveMux *http.ServeMux

	Manifest *PluginManifest
}

func MakePlugin() Plugin {
	r := Plugin{
		serveMux: http.NewServeMux(),
		Manifest: &PluginManifest{},
	}

	r.RegisterHandlerFunc(pluginActivatePath, r.handlePluginActivate)

	return r
}

func (z *Plugin) RegisterHandlerFunc(pattern string, handlerFn PluginHandlerFn) {
	util.Tracef("dockerSdkPlugin.Plugin.RegisterHandlerFunc(%s)\n", pattern)

	z.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", applicationDockerPluginsJsonMimeType)

		util.Tracef("%s %s plugin request\n", r.Method, r.URL.Path)

		if err := handlerFn(util.HttpRequest{Request: *r, ResponseWriter: w}); err != nil {
			util.Printf("Failed to process HTTP request %s %s: %v\n", r.Method, r.URL.Path, err)
		}
	})
}

func (z Plugin) handlePluginActivate(r util.HttpRequest) error {
	return r.WriteJson(z.Manifest)
}
