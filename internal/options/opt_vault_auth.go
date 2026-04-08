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
	VaultAuthMethodAppRole  = "approle"
	VaultAuthMethodCert     = "cert"
	VaultAuthMethodToken    = "token"
	VaultAuthMethodUserPass = "userpass"
)

var (
	clientDefaultAuthMountPathFromVaultAuthMethod = map[string]string{
		VaultAuthMethodAppRole:  "approle",
		VaultAuthMethodCert:     "cert",
		VaultAuthMethodToken:    "token",
		VaultAuthMethodUserPass: "userpass",
	}
)

type OptVaultAuth struct {
	Method        string  `json:","` // VaultAuthMethod*
	MountPath     *string `json:","`
	TokenRenewTtl int     `json:","`

	// AppRole
	RoleId               *string `json:","`
	RoleIdFile           *string `json:","`
	SecretId             *string `json:","`
	SecretIdFile         *string `json:","`
	SecretIdTokenWrapped bool    `json:","`

	// Cert
	CertFile    *string `json:","`
	CertKeyFile *string `json:","`

	// Token
	Token     *string `json:","`
	TokenFile *string `json:","`

	// Userpass
	Username     *string `json:","`
	UsernameFile *string `json:","`
	Password     *string `json:","`
	PasswordFile *string `json:","`
}

func (z OptVaultAuth) EffectiveMountPath() string {
	if z.MountPath == nil {
		v, ok := clientDefaultAuthMountPathFromVaultAuthMethod[z.Method]
		if ok {
			return v
		}

		return ""
	}

	return *z.MountPath
}

func (z OptVaultAuth) CacheId_() string {
	r := ""

	r += z.EffectiveMountPath() + z.Method
	r += strconv.Itoa(z.TokenRenewTtl)

	switch z.Method {
	case VaultAuthMethodAppRole:
		if z.RoleIdFile == nil {
			r += *z.RoleId
		} else {
			r += *z.RoleIdFile
		}
		if z.SecretIdFile == nil {
			r += *z.SecretId
		} else {
			r += *z.SecretIdFile
		}
	case VaultAuthMethodCert:
		r += *z.CertFile + *z.CertKeyFile
	case VaultAuthMethodToken:
		if z.TokenFile == nil {
			r += *z.Token
		} else {
			r += *z.TokenFile
		}
	case VaultAuthMethodUserPass:
		if z.UsernameFile == nil {
			r += *z.Username
		} else {
			r += *z.UsernameFile
		}
		if z.PasswordFile == nil {
			r += *z.Password
		} else {
			r += *z.PasswordFile
		}
	}

	return r
}

func MakeOptVaultAuth() OptVaultAuth {
	return OptVaultAuth{
		Method: VaultAuthMethodToken,
	}
}

func NewOptVaultAuthFromDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptVaultAuth) (*OptVaultAuth, error) {
	var r OptVaultAuth

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

func NewOptVaultAuthFromDockerSecret(secretName string, secretLabels map[string]string, serviceLabels map[string]string, defaultConfig *OptVaultAuth) (*OptVaultAuth, error) {
	var r OptVaultAuth

	if err := r.UpdateFromDockerSecret(secretName, secretLabels, serviceLabels); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptVaultAuth) UpdateFromDockerVolume(_ string, volumeOptions map[string]string) error {
	var v string
	var ok bool

	am, ok := volumeOptions["auth-mount"]
	if ok {
		z.MountPath = &am
	}
	v, ok = volumeOptions["auth-method"]
	if ok {
		z.Method = v
	}

	v, ok = volumeOptions["auth-token-renew-ttl"]
	if ok {
		atrt, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("unable to convert auth-token-renew-ttl value %s to integer", v)
		}

		z.TokenRenewTtl = atrt
	}

	voari, ok := volumeOptions["auth-role-id"]
	if ok {
		z.RoleId = &voari
	}
	voarif, ok := volumeOptions["auth-role-id-file"]
	if ok {
		z.RoleIdFile = &voarif
	}
	voasi, ok := volumeOptions["auth-secret-id"]
	if ok {
		z.SecretId = &voasi
	}
	voasif, ok := volumeOptions["auth-secret-id-file"]
	if ok {
		z.SecretIdFile = &voasif
	}

	voacf, ok := volumeOptions["auth-cert-file"]
	if ok {
		z.CertFile = &voacf
	}
	voackf, ok := volumeOptions["auth-cert-key-file"]
	if ok {
		z.CertKeyFile = &voackf
	}

	voat, ok := volumeOptions["auth-token"]
	if ok {
		z.Token = &voat
	}
	voatf, ok := volumeOptions["auth-token-file"]
	if ok {
		z.TokenFile = &voatf
	}

	voau, ok := volumeOptions["auth-username"]
	if ok {
		z.Username = &voau
	}
	voauf, ok := volumeOptions["auth-username-file"]
	if ok {
		z.UsernameFile = &voauf
	}
	voap, ok := volumeOptions["auth-password"]
	if ok {
		z.Password = &voap
	}
	voapf, ok := volumeOptions["auth-password-file"]
	if ok {
		z.PasswordFile = &voapf
	}

	return nil
}

func (z *OptVaultAuth) UpdateFromDockerSecret(_ string, _ map[string]string, _ map[string]string) error {
	return errors.New("not implemented")
}

func (z *OptVaultAuth) Normalize() {
	z.Method = strings.ToLower(z.Method)

	if z.MountPath != nil {
		if *z.MountPath == "" {
			z.MountPath = nil
		} else {
			v := path.Clean(*z.MountPath)
			z.MountPath = &v
		}
	}

	if z.RoleIdFile != nil && *z.RoleIdFile != "" {
		f := path.Clean(*z.RoleIdFile)

		z.RoleIdFile = &f
	}
	if z.SecretIdFile != nil && *z.SecretIdFile != "" {
		f := path.Clean(*z.SecretIdFile)

		z.SecretIdFile = &f
	}
	if z.CertFile != nil && *z.CertFile != "" {
		f := path.Clean(*z.CertFile)

		z.CertFile = &f
	}
	if z.CertKeyFile != nil && *z.CertKeyFile != "" {
		f := path.Clean(*z.CertKeyFile)

		z.CertKeyFile = &f
	}
	if z.TokenFile != nil && *z.TokenFile != "" {
		f := path.Clean(*z.TokenFile)

		z.TokenFile = &f
	}
	if z.UsernameFile != nil && *z.UsernameFile != "" {
		f := path.Clean(*z.UsernameFile)

		z.UsernameFile = &f
	}
	if z.PasswordFile != nil && *z.PasswordFile != "" {
		f := path.Clean(*z.PasswordFile)

		z.PasswordFile = &f
	}
}

func (z *OptVaultAuth) NormalizeAndValidate() error {
	z.Normalize()

	switch z.Method {
	case VaultAuthMethodAppRole:
		if z.RoleId == nil && z.RoleIdFile == nil {
			return errors.New("appRole auth method requires a RoleID to be defined")
		}
		if z.SecretId == nil && z.SecretIdFile == nil {
			return errors.New("appRole auth method requires a SecretID to be defined")
		}
	case VaultAuthMethodCert:
		if z.CertFile == nil || z.CertKeyFile == nil {
			return errors.New("cert auth method requires both cert and cert key files to be defined")
		}
	case VaultAuthMethodToken:
		if z.TokenFile == nil && z.Token == nil {
			return errors.New("token auth method requires a token to be defined")
		}
	case VaultAuthMethodUserPass:
		if z.Username == nil && z.UsernameFile == nil {
			return errors.New("userpass auth method requires an username to be defined")
		}
		if z.Password == nil && z.PasswordFile == nil {
			return errors.New("userpass auth method requires a password to be defined")
		}
	default:
		return fmt.Errorf("unknown auth method %s", z.Method)
	}

	return nil
}
