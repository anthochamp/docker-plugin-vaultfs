// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

type OptDocker struct {
	DockerVolume OptDockerVolume `json:","`
	Secret       OptSecret       `json:","`
}

func (z OptDocker) CacheId_() string {
	r := ""

	r += z.DockerVolume.CacheId_()
	r += z.Secret.CacheId_()

	return r
}

func MakeOptDocker() OptDocker {
	return OptDocker{
		DockerVolume: MakeOptDockerVolume(),
		Secret:       MakeOptSecret(),
	}
}

func NewOptDockerFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptDocker) (*OptDocker, error) {
	var r OptDocker

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

func NewOptDockerFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptDocker) (*OptDocker, error) {
	var r OptDocker

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptDocker) UpdateFromDockerVolume(volumeName string, volumeOptions map[string]string) error {
	if err := z.DockerVolume.Update(volumeName, volumeOptions); err != nil {
		return err
	}

	if err := z.Secret.UpdateFromDockerVolume(volumeName, volumeOptions); err != nil {
		return err
	}

	return nil
}

func (z *OptDocker) UpdateFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string) error {
	if err := z.Secret.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return err
	}

	return nil
}

func (z *OptDocker) Normalize() {
	z.DockerVolume.Normalize()
	z.Secret.Normalize()
}

func (z *OptDocker) NormalizeAndValidate() error {
	z.Normalize()

	if err := z.DockerVolume.NormalizeAndValidate(); err != nil {
		return err
	}

	if err := z.Secret.NormalizeAndValidate(); err != nil {
		return err
	}

	return nil
}
