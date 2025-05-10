// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"errors"

	"github.com/anthochamp/docker-plugin-vaultfs/backend"
	backendVault "github.com/anthochamp/docker-plugin-vaultfs/backend/vault"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
)

type SecretConfig struct {
	OptSecret options.OptSecret
}

func newSecret(config SecretConfig) (*backend.Secret, error) {
	var secret backend.Secret
	var err error

	switch config.OptSecret.Backend {
	case options.SecretBackendVault:
		secret, err = backendVault.NewVaultSecret(backendVault.VaultSecretConfig{
			OptVault: config.OptSecret.Vault,
		})

	default:
		return nil, errors.New("not implemented")
	}

	return &secret, err
}
