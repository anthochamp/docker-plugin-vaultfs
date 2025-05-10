// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"errors"
	"fmt"
	"net/url"
)

type OptClientHttp struct {
	Address          string       `json:","`
	DisableRedirects bool         `json:","`
	Tls              OptClientTls `json:","`
}

func (z OptClientHttp) CacheId_() string {
	r := ""

	r = z.Address

	if z.DisableRedirects {
		r += "1"
	} else {
		r += "0"
	}

	r += z.Tls.CacheId_()

	return r
}

func MakeOptClientHttp() OptClientHttp {
	return OptClientHttp{
		Tls: MakeOptClientTls(),
	}
}

func NewOptClientHttpFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptClientHttp) (*OptClientHttp, error) {
	var r OptClientHttp

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

func NewOptClientHttpFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptClientHttp) (*OptClientHttp, error) {
	var r OptClientHttp

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptClientHttp) UpdateFromDockerVolume(_ string, volumeOptions map[string]string) error {
	// TODO

	return nil
}

func (z *OptClientHttp) UpdateFromDockerSecret(_ string, _ map[string]string, _ map[string]string) error {
	return errors.New("not implemented")
}

func (z *OptClientHttp) Normalize() {
	z.Tls.Normalize()
}

func (z *OptClientHttp) NormalizeAndValidate() error {
	z.Normalize()

	if _, err := url.Parse(z.Address); err != nil {
		return fmt.Errorf("address is invalid: %w", err)
	}

	return nil
}
