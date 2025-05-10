// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package backendVault

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anthochamp/docker-plugin-vaultfs/constants"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/google/uuid"
	vaultApi "github.com/hashicorp/vault/api"
)

type VaultSecretData struct {
	secret *vaultApi.Secret

	uniqueId   string
	receivedAt time.Time
	cacheTtl   time.Duration

	data      map[string]string
	createdAt *time.Time
}

func (z VaultSecretData) UniqueId() string { return z.uniqueId }

func (z VaultSecretData) CreatedAt() *time.Time { return z.createdAt }

func (z VaultSecretData) GetKeys() []string {
	r := make([]string, 0, len(z.data))
	for k := range z.data {
		r = append(r, k)
	}
	return r
}
func (z VaultSecretData) GetValue(key string) (*string, bool) {
	value, ok := z.data[key]

	if !ok {
		return nil, false
	}

	return &value, true
}

func NewVaultSecretDataFromKVSecret(kvSecret vaultApi.KVSecret) (*VaultSecretData, error) {
	var createdAt *time.Time
	if kvSecret.VersionMetadata != nil {
		createdAt = &kvSecret.VersionMetadata.CreatedTime
	}

	var data map[string]string
	if kvSecret.Data == nil {
		data = map[string]string{}
	} else {
		data = util.MapStringStringFromMapStringInterface(kvSecret.Data)
	}

	if kvSecret.VersionMetadata != nil {
		data[".version-metadata-created-at"] = kvSecret.VersionMetadata.CreatedTime.UTC().Format(time.RFC3339)
		data[".version-metadata-deleted-at"] = kvSecret.VersionMetadata.DeletionTime.UTC().Format(time.RFC3339)
		data[".version-metadata-is-destroyed"] = strconv.FormatBool(kvSecret.VersionMetadata.Destroyed)
		data[".version-metadata-version"] = strconv.Itoa(kvSecret.VersionMetadata.Version)
	}

	// TODO: default cache ttl
	var cacheTtl time.Duration = 0

	if kvSecret.CustomMetadata != nil {
		for k, v := range kvSecret.CustomMetadata {
			if strings.HasPrefix(k, constants.AppName+"-") {
				if k == constants.AppName+"-cache-ttl" {
					cacheTtli, err := strconv.Atoi(v.(string))
					if err != nil {
						return nil, fmt.Errorf("unable to convert %s value %v to int", constants.AppName+"-cache-ttl", cacheTtli)
					}
					cacheTtl = time.Duration(cacheTtli)
				}
			} else {
				data[".metadata-"+k] = v.(string)
			}
		}
	}

	return &VaultSecretData{
		secret: kvSecret.Raw,

		uniqueId:   uuid.New().String(),
		receivedAt: time.Now(),
		cacheTtl:   cacheTtl,

		data:      data,
		createdAt: createdAt,
	}, nil
}
