// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
)

type OptVaultSecret struct {
	Path          string `json:","`
	TokenRenewTtl int    `json:","`

	KvVersion *int `json:","` // secret version for EngineKvVersion=2 (nil means "latest")
}

func (z OptVaultSecret) CacheId_() string {
	r := ""

	r = z.Path + strconv.Itoa(z.TokenRenewTtl)

	if z.KvVersion == nil {
		r += "nil"
	} else {
		r += strconv.Itoa(*z.KvVersion)
	}

	return r
}

func MakeOptVaultSecret() OptVaultSecret {
	return OptVaultSecret{}
}

func NewOptVaultSecretFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptVaultSecret) (*OptVaultSecret, error) {
	var r OptVaultSecret

	if defaultConfig != nil {
		r = *defaultConfig
	}

	if err := r.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func NewOptVaultSecretFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptVaultSecret) (*OptVaultSecret, error) {
	var r OptVaultSecret

	if defaultConfig != nil {
		r = *defaultConfig
	}

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptVaultSecret) UpdateFromDockerVolume(volumeName string, volumeOptions map[string]string) error {
	v, ok := volumeOptions["token-renew-ttl"]
	if ok {
		atrt, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("unable to convert token-renew-ttl value %s to integer", v)
		}

		z.TokenRenewTtl = atrt
	}

	vns := strings.SplitN(volumeName, "@", 2)

	vos, ok := volumeOptions["secret"]
	if ok {
		z.Path = vos
	} else {
		z.Path = vns[0]
	}

	ksv, ok := volumeOptions["kv-secret-version"]
	if !ok && len(vns) > 1 {
		if vns[1] == "" {
			ksv = "latest"
		} else {
			ksv = vns[1]
		}
		ok = true
	}

	if ok {
		if ksv != "latest" {
			v, err := strconv.Atoi(ksv)
			if err != nil {
				return fmt.Errorf("convert kv-secret-version %s to integer: %w", ksv, err)
			}

			z.KvVersion = &v
		} else {
			z.KvVersion = nil
		}
	}

	return nil
}

func (z *OptVaultSecret) UpdateFromDockerSecret(_ string, _ map[string]string, _ map[string]string) error {
	return errors.New("not implemented")
}

func (z *OptVaultSecret) NormalizeAndValidate() error {
	z.Normalize()

	if z.Path == "" {
		return errors.New("path cannot be empty")
	}

	return nil
}

func (z *OptVaultSecret) Normalize() {
	if z.Path != "" {
		z.Path = path.Clean(z.Path)
	}
}
