// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package backendVault

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/anthochamp/docker-plugin-vaultfs/backend"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	vaultApi "github.com/hashicorp/vault/api"
)

type VaultSecret struct {
	optVaultEngine options.OptVaultEngine
	optVaultSecret options.OptVaultSecret
	client         *VaultClient

	cacheLock         *sync.Mutex
	lifetimeWatcherId *string
	cacheRefTime      time.Time
	cacheTtl          time.Duration
	data              *backend.SecretData
}

type VaultSecretConfig struct {
	OptVault options.OptVault
}

func NewVaultSecret(config VaultSecretConfig) (*VaultSecret, error) {
	client, err := newVaultClient(VaultClientConfig{
		optClientHttp: config.OptVault.ClientHttp,
		optVaultAuth:  config.OptVault.VaultAuth,
	})
	if err != nil {
		return nil, fmt.Errorf("create vault client: %w", err)
	}

	return &VaultSecret{
		optVaultEngine: config.OptVault.VaultEngine,
		optVaultSecret: config.OptVault.VaultSecret,
		client:         client,

		cacheLock: &sync.Mutex{},
	}, nil
}

func (z *VaultSecret) Close() {
	z.cacheLock.Lock()
	z.clearCacheUnsafe()
	z.cacheLock.Unlock()

	z.client.Close()
	z.client = nil
}

func (z *VaultSecret) GetData(noCache bool) (*backend.SecretData, error) {
	util.Tracef("VaultSecret[%v].GetData(%v)\n", z, noCache)

	z.cacheLock.Lock()
	defer z.cacheLock.Unlock()

	if !noCache && z.data != nil && (z.cacheTtl == 0 || time.Now().Compare(z.cacheRefTime.Add(z.cacheTtl)) < 0) {
		return z.data, nil
	}

	z.clearCacheUnsafe()

	var data *VaultSecretData
	var err error

	switch z.optVaultEngine.Type {
	case options.VaultEngineTypeKv:
		data, err = z.getKvData()

	default:
		return nil, errors.New("not implemented")
	}

	if err != nil {
		return nil, err
	}

	if len(data.secret.Warnings) > 0 {
		util.Noticef("Received vault secret has warning: %v", data.secret.Warnings)
	}

	if data.secret.Renewable {
		lifetimeWatcherId, err := z.client.NewLifetimeWatcher(vaultApi.LifetimeWatcherInput{
			Secret:    data.secret,
			Increment: z.optVaultSecret.TokenRenewTtl,
		}, func(err error) {
			if err != nil {
				util.Errorf("Unable to renew vault data secret %v: %v\n", z, err)
			}

			z.clearCacheUnsafe()
		}, func(renewal *vaultApi.RenewOutput) {
			util.Tracef("Renewed vault data secret %v: %+v\n", z, renewal)
		})

		z.lifetimeWatcherId = lifetimeWatcherId

		if err != nil {
			return nil, fmt.Errorf("create lifetime watcher: %w", err)
		}
	}

	var _data backend.SecretData = *data
	z.data = &_data
	z.cacheRefTime = data.receivedAt
	z.cacheTtl = data.cacheTtl

	return z.data, nil
}

func (z *VaultSecret) getKvData() (*VaultSecretData, error) {
	var kvSecret *vaultApi.KVSecret
	var err error

	switch z.optVaultEngine.KvVersion {
	case 1:
		kvSecret, err = z.client.FetchKVv1Secret(z.optVaultEngine.EffectiveMountPath(), z.optVaultSecret.Path)
	case 2:
		kvSecret, err = z.client.FetchKVv2Secret(z.optVaultEngine.EffectiveMountPath(), z.optVaultSecret.Path, z.optVaultSecret.KvVersion)
	}

	if err != nil {
		return nil, err
	}

	data, err := NewVaultSecretDataFromKVSecret(*kvSecret)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (z *VaultSecret) clearCacheUnsafe() {
	if z.lifetimeWatcherId != nil {
		// avoids deadlock when closing
		id := z.lifetimeWatcherId
		z.lifetimeWatcherId = nil

		z.client.CloseLifetimeWatcher(*id)
	}

}
