// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"testing"
)

func TestOptVaultAuthNormalizeAndValidate(t *testing.T) {
	t.Run("token method valid with token set", func(t *testing.T) {
		token := "s.mytoken"
		opt := OptVaultAuth{Method: VaultAuthMethodToken, Token: &token}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("token method valid with token file set", func(t *testing.T) {
		tokenFile := "/run/secrets/vault-token"
		opt := OptVaultAuth{Method: VaultAuthMethodToken, TokenFile: &tokenFile}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("token method missing token returns error", func(t *testing.T) {
		opt := OptVaultAuth{Method: VaultAuthMethodToken}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when token is not set")
		}
	})

	t.Run("approle method valid with role-id and secret-id", func(t *testing.T) {
		roleId := "my-role-id"
		secretId := "my-secret-id"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodAppRole,
			RoleId:   &roleId,
			SecretId: &secretId,
		}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("approle method valid with file-based credentials", func(t *testing.T) {
		roleIdFile := "/run/secrets/role-id"
		secretIdFile := "/run/secrets/secret-id"
		opt := OptVaultAuth{
			Method:       VaultAuthMethodAppRole,
			RoleIdFile:   &roleIdFile,
			SecretIdFile: &secretIdFile,
		}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("approle method missing role-id returns error", func(t *testing.T) {
		secretId := "my-secret-id"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodAppRole,
			SecretId: &secretId,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when role-id is not set")
		}
	})

	t.Run("approle method missing secret-id returns error", func(t *testing.T) {
		roleId := "my-role-id"
		opt := OptVaultAuth{
			Method: VaultAuthMethodAppRole,
			RoleId: &roleId,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when secret-id is not set")
		}
	})

	t.Run("cert method valid with cert and key files", func(t *testing.T) {
		certFile := "/certs/client.crt"
		certKeyFile := "/certs/client.key"
		opt := OptVaultAuth{
			Method:      VaultAuthMethodCert,
			CertFile:    &certFile,
			CertKeyFile: &certKeyFile,
		}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("cert method missing cert file returns error", func(t *testing.T) {
		certKeyFile := "/certs/client.key"
		opt := OptVaultAuth{
			Method:      VaultAuthMethodCert,
			CertKeyFile: &certKeyFile,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when cert file is not set")
		}
	})

	t.Run("cert method missing cert key file returns error", func(t *testing.T) {
		certFile := "/certs/client.crt"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodCert,
			CertFile: &certFile,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when cert key file is not set")
		}
	})

	t.Run("userpass method valid with username and password", func(t *testing.T) {
		username := "alice"
		password := "correct-horse-battery-staple"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodUserPass,
			Username: &username,
			Password: &password,
		}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("userpass method missing username returns error", func(t *testing.T) {
		password := "correct-horse-battery-staple"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodUserPass,
			Password: &password,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when username is not set")
		}
	})

	t.Run("userpass method missing password returns error", func(t *testing.T) {
		username := "alice"
		opt := OptVaultAuth{
			Method:   VaultAuthMethodUserPass,
			Username: &username,
		}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error when password is not set")
		}
	})

	t.Run("unknown method returns error", func(t *testing.T) {
		opt := OptVaultAuth{Method: "unknown-method"}

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error for unknown auth method")
		}
	})

	t.Run("method is normalized to lowercase before validation", func(t *testing.T) {
		token := "s.mytoken"
		opt := OptVaultAuth{Method: "TOKEN", Token: &token}

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if opt.Method != VaultAuthMethodToken {
			t.Errorf("expected method %q after normalization, got %q", VaultAuthMethodToken, opt.Method)
		}
	})
}

func TestOptVaultAuthEffectiveMountPath(t *testing.T) {
	t.Run("returns method default when mount path is not set", func(t *testing.T) {
		tests := []struct {
			method   string
			expected string
		}{
			{VaultAuthMethodAppRole, "approle"},
			{VaultAuthMethodCert, "cert"},
			{VaultAuthMethodToken, "token"},
			{VaultAuthMethodUserPass, "userpass"},
		}

		for _, tc := range tests {
			opt := OptVaultAuth{Method: tc.method}

			got := opt.EffectiveMountPath()

			if got != tc.expected {
				t.Errorf("method %q: expected default mount path %q, got %q", tc.method, tc.expected, got)
			}
		}
	})

	t.Run("returns explicit mount path when set", func(t *testing.T) {
		mountPath := "custom/approle"
		opt := OptVaultAuth{
			Method:    VaultAuthMethodAppRole,
			MountPath: &mountPath,
		}

		if opt.EffectiveMountPath() != mountPath {
			t.Errorf("expected %q, got %q", mountPath, opt.EffectiveMountPath())
		}
	})
}

func TestOptVaultAuthUpdateFromDockerVolume(t *testing.T) {
	t.Run("auth-method option sets the method", func(t *testing.T) {
		opt := MakeOptVaultAuth()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"auth-method": "approle"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Method != "approle" {
			t.Errorf("expected method %q, got %q", "approle", opt.Method)
		}
	})

	t.Run("auth-mount option sets the mount path", func(t *testing.T) {
		opt := MakeOptVaultAuth()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"auth-mount": "my/approle"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountPath == nil || *opt.MountPath != "my/approle" {
			t.Errorf("expected mount path %q, got %v", "my/approle", opt.MountPath)
		}
	})

	t.Run("auth-token-renew-ttl valid value is parsed", func(t *testing.T) {
		opt := MakeOptVaultAuth()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"auth-token-renew-ttl": "300"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.TokenRenewTtl != 300 {
			t.Errorf("expected TokenRenewTtl=300, got %d", opt.TokenRenewTtl)
		}
	})

	t.Run("auth-token-renew-ttl invalid value returns error", func(t *testing.T) {
		opt := MakeOptVaultAuth()

		err := opt.UpdateFromDockerVolume("vol", map[string]string{"auth-token-renew-ttl": "bad"})

		if err == nil {
			t.Error("expected error for invalid auth-token-renew-ttl")
		}
	})
}
