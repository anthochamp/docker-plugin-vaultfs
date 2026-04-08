// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPluginActivate(t *testing.T) {
	t.Run("returns empty implements on fresh plugin", func(t *testing.T) {
		plugin := MakePlugin()

		request := httptest.NewRequest(http.MethodPost, pluginActivatePath, nil)
		recorder := httptest.NewRecorder()

		plugin.serveMux.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		contentType := recorder.Header().Get("Content-Type")
		if contentType != applicationDockerPluginsJsonMimeType {
			t.Errorf("expected Content-Type %q, got %q", applicationDockerPluginsJsonMimeType, contentType)
		}

		var manifest PluginManifest
		if err := json.NewDecoder(recorder.Body).Decode(&manifest); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(manifest.Implements) != 0 {
			t.Errorf("expected empty Implements, got %v", manifest.Implements)
		}
	})

	t.Run("after registering VolumeDriver Implements contains VolumeDriver id", func(t *testing.T) {
		plugin := MakePlugin()
		mockDriver := &mockVolumeDriver{}
		RegisterVolumeDriver(mockDriver, plugin)

		request := httptest.NewRequest(http.MethodPost, pluginActivatePath, nil)
		recorder := httptest.NewRecorder()

		plugin.serveMux.ServeHTTP(recorder, request)

		var manifest PluginManifest
		if err := json.NewDecoder(recorder.Body).Decode(&manifest); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(manifest.Implements) != 1 || manifest.Implements[0] != pluginManifestVolumeDriverImplementId {
			t.Errorf("expected Implements=[%q], got %v", pluginManifestVolumeDriverImplementId, manifest.Implements)
		}
	})

	t.Run("after registering SecretProvider Implements contains secretprovider id", func(t *testing.T) {
		plugin := MakePlugin()
		mockProvider := &mockSecretProvider{}
		RegisterSecretProvider(mockProvider, plugin)

		request := httptest.NewRequest(http.MethodPost, pluginActivatePath, nil)
		recorder := httptest.NewRecorder()

		plugin.serveMux.ServeHTTP(recorder, request)

		var manifest PluginManifest
		if err := json.NewDecoder(recorder.Body).Decode(&manifest); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(manifest.Implements) != 1 || manifest.Implements[0] != pluginManifestSecretProviderImplementId {
			t.Errorf("expected Implements=[%q], got %v", pluginManifestSecretProviderImplementId, manifest.Implements)
		}
	})
}
