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

// OptSecret holds configuration for a secret backend and its options.
type OptSecret struct {
	Backend string   `json:","`
	Vault   OptVault `json:","`
}

// CacheId_ returns a unique string representing the secret config for caching purposes.
func (z OptSecret) CacheId_() string {
	r := ""

	r += z.Backend

	switch z.Backend {
	case SecretBackendVault:
		r += z.Vault.CacheId_()
	}

	return r
}

// MakeOptSecret returns a new OptSecret with default values.
func MakeOptSecret() OptSecret {
	return OptSecret{
		Backend: SecretBackendVault,
		Vault:   MakeOptVault(),
	}
}

// NewOptSecretFromDockerVolume creates an OptSecret from Docker volume options.
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

// NewOptSecretFromDockerSecret creates an OptSecret from Docker secret and service labels.
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

// UpdateFromDockerVolume updates the OptSecret from Docker volume options.
func (z *OptSecret) UpdateFromDockerVolume(volumeName string, volumeOptions map[string]string) error {
	// TODO: backend

	if err := z.Vault.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	return nil
}

// UpdateFromDockerSecret updates the OptSecret from Docker secret and service labels.
func (z *OptSecret) UpdateFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string) error {
	// TODO: backend

	if err := z.Vault.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	return nil
}

// Normalize cleans up and standardizes the OptSecret fields.
func (z *OptSecret) Normalize() {
	// TODO: backend

	z.Vault.Normalize()
}

// NormalizeAndValidate normalizes and validates the OptSecret fields.
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
