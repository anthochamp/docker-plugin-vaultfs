// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"fmt"

	dockerSdkPlugin "github.com/anthochamp/docker-plugin-vaultfs/dockersdk/plugin"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type SecretProvider struct {
	SecretProviderConfig
}

type SecretProviderConfig struct {
	DefaultOptDocker options.OptDocker
}

func NewSecretProvider(config SecretProviderConfig) (*SecretProvider, error) {
	return &SecretProvider{
		SecretProviderConfig: config,
	}, nil
}

/***/

func (z SecretProvider) GetSecret(r dockerSdkPlugin.SecretProviderGetSecretRequest) (*dockerSdkPlugin.SecretProviderGetSecretResponse, error) {
	util.Tracef("SecretProvider.Get(%+v)\n", r)

	optDocker, err := options.NewOptDockerFromDockerSecret(r.SecretName, r.SecretLabels, r.ServiceLabels, &z.DefaultOptDocker)
	if err != nil {
		return nil, err
	}

	secret, err := newSecret(SecretConfig{
		OptSecret: optDocker.Secret,
	})
	if err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}

	defer (*secret).Close()

	data, err := (*secret).GetData(false)
	if err != nil {
		return nil, fmt.Errorf("get secret data: %w", err)
	}

	value, ok := (*data).GetValue(r.SecretName)
	if !ok {
		return nil, fmt.Errorf("get secret data field %s: %w", r.SecretName, err)
	}

	return &dockerSdkPlugin.SecretProviderGetSecretResponse{
		DoNotReuse: true,
		Value:      []byte(*value),
	}, nil
}
