// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

type OptVault struct {
	ClientHttp  OptClientHttp  `json:","`
	VaultAuth   OptVaultAuth   `json:","`
	VaultEngine OptVaultEngine `json:","`
	VaultSecret OptVaultSecret `json:","`
}

func (z OptVault) CacheId_() string {
	r := ""

	r += z.ClientHttp.CacheId_()
	r += z.VaultAuth.CacheId_()
	r += z.VaultEngine.CacheId_()
	r += z.VaultSecret.CacheId_()

	return r
}

func MakeOptVault() OptVault {
	return OptVault{
		ClientHttp:  MakeOptClientHttp(),
		VaultAuth:   MakeOptVaultAuth(),
		VaultEngine: MakeOptVaultEngine(),
		VaultSecret: MakeOptVaultSecret(),
	}
}

func NewOptVaultFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptVault) (*OptVault, error) {
	var r OptVault

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

func NewOptVaultFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptVault) (*OptVault, error) {
	var r OptVault

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptVault) UpdateFromDockerVolume(volumeName string, volumeOptions map[string]string) error {
	if err := z.ClientHttp.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	if err := z.VaultAuth.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	if err := z.VaultEngine.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	if err := z.VaultSecret.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	return nil
}

func (z *OptVault) UpdateFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string) error {
	if err := z.ClientHttp.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	if err := z.VaultAuth.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	if err := z.VaultEngine.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	if err := z.VaultSecret.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	return nil
}

func (z *OptVault) Normalize() {
	z.ClientHttp.Normalize()
	z.VaultAuth.Normalize()
	z.VaultEngine.Normalize()
	z.VaultSecret.Normalize()
}

func (z *OptVault) NormalizeAndValidate() error {
	z.Normalize()

	if err := z.ClientHttp.NormalizeAndValidate(); err != nil {
		return err
	}

	if err := z.VaultAuth.NormalizeAndValidate(); err != nil {
		return err
	}

	if err := z.VaultEngine.NormalizeAndValidate(); err != nil {
		return err
	}

	if err := z.VaultSecret.NormalizeAndValidate(); err != nil {
		return err
	}

	return nil
}
