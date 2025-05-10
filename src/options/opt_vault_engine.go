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

const (
	VaultEngineTypeKv  = "kv"
	VaultEngineTypeDb  = "db"
	VaultEngineTypePki = "pki"
)

var (
	VaultEngineTypes = []string{
		VaultEngineTypeKv,
		VaultEngineTypeDb,
		VaultEngineTypePki,
	}

	VaultEngineDefaultMountPathFromType = map[string]string{
		VaultEngineTypeKv:  "secret",
		VaultEngineTypeDb:  "database",
		VaultEngineTypePki: "pki",
	}
)

type OptVaultEngine struct {
	Type      string  `json:","` // VaultEngineType*
	MountPath *string `json:","`
	KvVersion int     `json:","`
}

func (z OptVaultEngine) CacheId_() string {
	r := ""

	r += z.Type

	if z.MountPath == nil {
		r += "nil"
	} else {
		r += *z.MountPath
	}

	r += strconv.Itoa(int(z.KvVersion))

	return r
}

func (z OptVaultEngine) EffectiveMountPath() string {
	if z.MountPath == nil {
		v, ok := VaultEngineDefaultMountPathFromType[z.Type]
		if ok {
			return v
		}

		return ""
	}

	return *z.MountPath
}

func MakeOptVaultEngine() OptVaultEngine {
	return OptVaultEngine{
		Type:      VaultEngineTypeKv,
		KvVersion: 1,
	}
}

func NewOptVaultEngineFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptVaultEngine) (*OptVaultEngine, error) {
	var r OptVaultEngine

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

func NewOptVaultEngineFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptVaultEngine) (*OptVaultEngine, error) {
	var r OptVaultEngine

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

func (z *OptVaultEngine) UpdateFromDockerVolume(_ string, volumeOptions map[string]string) error {
	var v string
	var ok bool

	v, ok = volumeOptions["engine-type"]
	if ok {
		z.Type = v
	}

	voem, ok := volumeOptions["engine-mount"]
	if ok {
		z.MountPath = &voem
	}

	v, ok = volumeOptions["kv-engine-version"]
	if ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("convert kv-engine-version %s to integer: %w", v, err)
		}

		z.KvVersion = i
	}

	return nil
}

func (z *OptVaultEngine) UpdateFromDockerSecret(_ string, _ map[string]string, _ map[string]string) error {
	return errors.New("not implemented")
}

func (z *OptVaultEngine) Normalize() {
	z.Type = strings.ToLower(z.Type)

	if z.MountPath != nil {
		if *z.MountPath == "" {
			z.MountPath = nil
		} else {
			v := path.Clean(*z.MountPath)
			z.MountPath = &v
		}
	}
}

func (z *OptVaultEngine) NormalizeAndValidate() error {
	z.Normalize()

	if z.MountPath != nil && *z.MountPath == "" {
		return errors.New("mount path cannot be empty")
	}

	switch z.Type {
	case VaultEngineTypeKv:
		if z.KvVersion != 1 && z.KvVersion != 2 {
			return fmt.Errorf("unknown KV version %d", z.KvVersion)
		}

	case VaultEngineTypeDb:
	case VaultEngineTypePki:
	default:
		return fmt.Errorf("unknown type %s", z.Type)
	}

	return nil
}
