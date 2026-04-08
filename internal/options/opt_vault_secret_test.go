// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"testing"
)

func TestOptVaultSecretUpdateFromDockerVolume(t *testing.T) {
	t.Run("volume name is used as secret path", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		if err := opt.UpdateFromDockerVolume("my/secret", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Path != "my/secret" {
			t.Errorf("expected path %q, got %q", "my/secret", opt.Path)
		}
	})

	t.Run("secret option overrides volume name as path", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		if err := opt.UpdateFromDockerVolume("ignored-name", map[string]string{"secret": "override/path"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Path != "override/path" {
			t.Errorf("expected path %q, got %q", "override/path", opt.Path)
		}
	})

	t.Run("at-sign suffix sets KV version", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		if err := opt.UpdateFromDockerVolume("my/secret@3", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Path != "my/secret" {
			t.Errorf("expected path %q, got %q", "my/secret", opt.Path)
		}

		if opt.KvVersion == nil || *opt.KvVersion != 3 {
			t.Errorf("expected KvVersion=3, got %v", opt.KvVersion)
		}
	})

	t.Run("at-latest suffix sets KV version to nil", func(t *testing.T) {
		version := 1
		opt := MakeOptVaultSecret()
		opt.KvVersion = &version

		if err := opt.UpdateFromDockerVolume("my/secret@latest", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.KvVersion != nil {
			t.Errorf("expected KvVersion=nil, got %d", *opt.KvVersion)
		}
	})

	t.Run("empty at-sign suffix sets KV version to nil", func(t *testing.T) {
		version := 1
		opt := MakeOptVaultSecret()
		opt.KvVersion = &version

		if err := opt.UpdateFromDockerVolume("my/secret@", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.KvVersion != nil {
			t.Errorf("expected KvVersion=nil, got %d", *opt.KvVersion)
		}
	})

	t.Run("kv-secret-version option takes precedence over at-sign suffix", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		if err := opt.UpdateFromDockerVolume("my/secret@1", map[string]string{"kv-secret-version": "5"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.KvVersion == nil || *opt.KvVersion != 5 {
			t.Errorf("expected KvVersion=5, got %v", opt.KvVersion)
		}
	})

	t.Run("kv-secret-version latest sets KV version to nil", func(t *testing.T) {
		version := 2
		opt := MakeOptVaultSecret()
		opt.KvVersion = &version

		if err := opt.UpdateFromDockerVolume("my/secret", map[string]string{"kv-secret-version": "latest"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.KvVersion != nil {
			t.Errorf("expected KvVersion=nil, got %d", *opt.KvVersion)
		}
	})

	t.Run("token-renew-ttl option is parsed as integer", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		if err := opt.UpdateFromDockerVolume("my/secret", map[string]string{"token-renew-ttl": "120"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.TokenRenewTtl != 120 {
			t.Errorf("expected TokenRenewTtl=120, got %d", opt.TokenRenewTtl)
		}
	})

	t.Run("invalid token-renew-ttl returns error", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		err := opt.UpdateFromDockerVolume("my/secret", map[string]string{"token-renew-ttl": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid token-renew-ttl")
		}
	})

	t.Run("invalid kv-secret-version returns error", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		err := opt.UpdateFromDockerVolume("my/secret", map[string]string{"kv-secret-version": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid kv-secret-version")
		}
	})
}

func TestOptVaultSecretNormalizeAndValidate(t *testing.T) {
	t.Run("empty path returns error", func(t *testing.T) {
		opt := MakeOptVaultSecret()

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("non-empty path passes validation", func(t *testing.T) {
		opt := MakeOptVaultSecret()
		opt.Path = "some/secret"

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("path is cleaned on normalize", func(t *testing.T) {
		opt := MakeOptVaultSecret()
		opt.Path = "some//secret/../secret"

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Path != "some/secret" {
			t.Errorf("expected cleaned path %q, got %q", "some/secret", opt.Path)
		}
	})
}
