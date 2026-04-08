// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockSecretProvider implements SecretProvider with configurable per-call behaviour.
type mockSecretProvider struct {
	getSecretFn func(SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error)
}

func (m *mockSecretProvider) GetSecret(req SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error) {
	if m.getSecretFn != nil {
		return m.getSecretFn(req)
	}
	return &SecretProviderGetSecretResponse{}, nil
}

// doSecretProviderRequest sends a POST request to the GetSecret path on a plugin
// that has the provided SecretProvider registered.
func doSecretProviderRequest(provider SecretProvider, body string) *httptest.ResponseRecorder {
	plugin := MakePlugin()
	RegisterSecretProvider(provider, plugin)

	request := httptest.NewRequest(http.MethodPost, secretProviderGetSecretPath, strings.NewReader(body))
	recorder := httptest.NewRecorder()

	plugin.serveMux.ServeHTTP(recorder, request)

	return recorder
}

func TestSecretProviderGetSecret(t *testing.T) {
	t.Run("happy path returns value and DoNotReuse", func(t *testing.T) {
		provider := &mockSecretProvider{
			getSecretFn: func(_ SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error) {
				return &SecretProviderGetSecretResponse{
					Value:      []byte("top-secret"),
					DoNotReuse: true,
				}, nil
			},
		}

		recorder := doSecretProviderRequest(provider, `{"SecretName":"my-secret"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response secretProviderGetSecretHttpResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if string(response.Value) != "top-secret" {
			t.Errorf("expected Value=%q, got %q", "top-secret", string(response.Value))
		}

		if !response.DoNotReuse {
			t.Error("expected DoNotReuse=true")
		}
	})

	t.Run("provider error returns 500 with Err field and no Value", func(t *testing.T) {
		provider := &mockSecretProvider{
			getSecretFn: func(_ SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error) {
				return nil, errors.New("vault unavailable")
			},
		}

		recorder := doSecretProviderRequest(provider, `{"SecretName":"my-secret"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}

		var response secretProviderGetSecretHttpResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Err != "vault unavailable" {
			t.Errorf("expected Err=%q, got %q", "vault unavailable", response.Err)
		}

		if len(response.Value) != 0 {
			t.Errorf("expected no Value on error, got %q", string(response.Value))
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		recorder := doSecretProviderRequest(&mockSecretProvider{}, `{invalid}`)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("secret name is forwarded to the provider", func(t *testing.T) {
		var receivedName string

		provider := &mockSecretProvider{
			getSecretFn: func(req SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error) {
				receivedName = req.SecretName
				return &SecretProviderGetSecretResponse{Value: []byte("value")}, nil
			},
		}

		doSecretProviderRequest(provider, `{"SecretName":"db-password"}`)

		if receivedName != "db-password" {
			t.Errorf("expected SecretName=%q to be forwarded, got %q", "db-password", receivedName)
		}
	})
}
