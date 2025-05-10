// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

const (
	SecretBackendVault = "vault"
)

var (
	SecretBackends = []string{
		SecretBackendVault,
	}
)

type OptSecret struct {
	Backend string   `json:","`
	Vault   OptVault `json:","`
}

func (z OptSecret) CacheId_() string {
	r := ""

	r += z.Backend

	switch z.Backend {
	case SecretBackendVault:
		r += z.Vault.CacheId_()
	}

	return r
}

func MakeOptSecret() OptSecret {
	return OptSecret{
		Backend: SecretBackendVault,
		Vault:   MakeOptVault(),
	}
}

func NewOptSecretFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptSecret) (*OptSecret, error) {
	var r OptSecret

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

func NewOptSecretFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptSecret) (*OptSecret, error) {
	var r OptSecret

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptSecret) UpdateFromDockerVolume(volumeName string, volumeOptions map[string]string) error {
	// TODO: backend

	if err := z.Vault.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	return nil
}

func (z *OptSecret) UpdateFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string) error {
	// TODO: backend

	if err := z.Vault.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	return nil
}

func (z *OptSecret) Normalize() {
	// TODO: backend

	z.Vault.Normalize()
}

func (z *OptSecret) NormalizeAndValidate() error {
	z.Normalize()

	switch z.Backend {
	case SecretBackendVault:
		if err := z.Vault.NormalizeAndValidate(); err != nil {
			return err
		}
	}

	return nil
}
