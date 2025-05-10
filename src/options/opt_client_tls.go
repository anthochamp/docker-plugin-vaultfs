// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"errors"
	"path"
)

type OptClientTls struct {
	Insecure   bool    `json:","`
	CACertFile *string `json:","`
	CertFile   *string `json:","`
	KeyFile    *string `json:","`
	ServerName *string `json:","`
}

func (z OptClientTls) CacheId_() string {
	r := ""

	if z.Insecure {
		r += "1"
	} else {
		r += "0"
	}

	if z.CACertFile != nil {
		r += *z.CACertFile
	} else {
		r += "nil"
	}

	if z.CertFile != nil {
		r += *z.CertFile
	} else {
		r += "nil"
	}

	if z.KeyFile != nil {
		r += *z.KeyFile
	} else {
		r += "nil"
	}

	if z.ServerName != nil {
		r += *z.ServerName
	} else {
		r += "nil"
	}

	return r
}

func MakeOptClientTls() OptClientTls {
	return OptClientTls{}
}

func NewOptClientTlsFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptClientTls) (*OptClientTls, error) {
	var r OptClientTls

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

func NewOptClientTlsFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptClientTls) (*OptClientTls, error) {
	var r OptClientTls

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptClientTls) UpdateFromDockerVolume(_ string, volumeOptions map[string]string) error {
	// TODO

	return nil
}

func (z *OptClientTls) UpdateFromDockerSecret(_ string, _ map[string]string, _ map[string]string) error {
	return errors.New("not implemented")
}

func (z *OptClientTls) Normalize() {
	if z.CACertFile != nil {
		if *z.CACertFile == "" {
			z.CACertFile = nil
		} else {
			v := path.Clean(*z.CACertFile)
			z.CACertFile = &v
		}
	}

	if z.CertFile != nil {
		if *z.CertFile == "" {
			z.CertFile = nil
		} else {
			v := path.Clean(*z.CertFile)
			z.CertFile = &v
		}
	}

	if z.KeyFile != nil {
		if *z.KeyFile == "" {
			z.KeyFile = nil
		} else {
			v := path.Clean(*z.KeyFile)
			z.KeyFile = &v
		}
	}

	if z.ServerName != nil && *z.ServerName == "" {
		z.ServerName = nil
	}
}

func (z *OptClientTls) NormalizeAndValidate() error {
	z.Normalize()

	return nil
}
