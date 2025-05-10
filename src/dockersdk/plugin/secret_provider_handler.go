// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"net/http"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type SecretProvider interface {
	GetSecret(SecretProviderGetSecretRequest) (*SecretProviderGetSecretResponse, error)
}

func RegisterSecretProvider(sp SecretProvider, p Plugin) {
	p.RegisterHandlerFunc(secretProviderGetSecretPath, func(hr util.HttpRequest) error { return handleSecretProviderGetSecret(sp, hr) })

	p.Manifest.Implements = append(p.Manifest.Implements, pluginManifestSecretProviderImplementId)
}

func handleSecretProviderGetSecret(sp SecretProvider, hr util.HttpRequest) error {
	var req SecretProviderGetSecretRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	res, err := sp.GetSecret(req)

	var httpRes secretProviderGetSecretHttpResponse

	if err != nil {
		hr.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		httpRes.Err = err.Error()
	}

	if res != nil {
		httpRes.DoNotReuse = res.DoNotReuse
		httpRes.Value = res.Value
	}

	return hr.WriteJson(httpRes)
}
