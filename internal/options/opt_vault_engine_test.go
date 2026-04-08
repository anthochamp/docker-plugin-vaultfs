// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"testing"
)

func TestOptVaultEngineNormalizeAndValidate(t *testing.T) {
	t.Run("default type kv with version 1 is valid", func(t *testing.T) {
		opt := MakeOptVaultEngine()

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("kv type version 2 is valid", func(t *testing.T) {
		opt := MakeOptVaultEngine()
		opt.KvVersion = 2

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("kv type unsupported version returns error", func(t *testing.T) {
		tests := []int{0, 3, 10}

		for _, version := range tests {
			opt := MakeOptVaultEngine()
			opt.KvVersion = version

			if err := opt.NormalizeAndValidate(); err == nil {
				t.Errorf("expected error for kv version %d", version)
			}
		}
	})

	t.Run("db type is valid", func(t *testing.T) {
		opt := MakeOptVaultEngine()
		opt.Type = VaultEngineTypeDb

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("pki type is valid", func(t *testing.T) {
		opt := MakeOptVaultEngine()
		opt.Type = VaultEngineTypePki

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unknown type returns error", func(t *testing.T) {
		opt := MakeOptVaultEngine()
		opt.Type = "unknown-engine"

		err := opt.NormalizeAndValidate()

		if err == nil {
			t.Error("expected error for unknown engine type")
		}
	})

	t.Run("type is normalized to lowercase before validation", func(t *testing.T) {
		opt := MakeOptVaultEngine()
		opt.Type = "KV"

		if err := opt.NormalizeAndValidate(); err != nil {
			t.Errorf("unexpected error after lowercase normalization: %v", err)
		}

		if opt.Type != VaultEngineTypeKv {
			t.Errorf("expected type %q after normalization, got %q", VaultEngineTypeKv, opt.Type)
		}
	})
}

func TestOptVaultEngineEffectiveMountPath(t *testing.T) {
	t.Run("returns type default when mount path is not set", func(t *testing.T) {
		tests := []struct {
			engineType string
			expected   string
		}{
			{VaultEngineTypeKv, "secret"},
			{VaultEngineTypeDb, "database"},
			{VaultEngineTypePki, "pki"},
		}

		for _, tc := range tests {
			opt := OptVaultEngine{Type: tc.engineType}

			got := opt.EffectiveMountPath()

			if got != tc.expected {
				t.Errorf("type %q: expected default mount path %q, got %q", tc.engineType, tc.expected, got)
			}
		}
	})

	t.Run("returns explicit mount path when set", func(t *testing.T) {
		mountPath := "my-kv"
		opt := OptVaultEngine{
			Type:      VaultEngineTypeKv,
			MountPath: &mountPath,
		}

		if opt.EffectiveMountPath() != mountPath {
			t.Errorf("expected %q, got %q", mountPath, opt.EffectiveMountPath())
		}
	})
}

func TestOptVaultEngineUpdateFromDockerVolume(t *testing.T) {
	t.Run("engine-type option sets the type", func(t *testing.T) {
		opt := MakeOptVaultEngine()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"engine-type": "db"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.Type != "db" {
			t.Errorf("expected type %q, got %q", "db", opt.Type)
		}
	})

	t.Run("engine-mount option sets the mount path", func(t *testing.T) {
		opt := MakeOptVaultEngine()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"engine-mount": "my-secrets"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountPath == nil || *opt.MountPath != "my-secrets" {
			t.Errorf("expected mount path %q, got %v", "my-secrets", opt.MountPath)
		}
	})

	t.Run("kv-engine-version option sets the KV version", func(t *testing.T) {
		opt := MakeOptVaultEngine()

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"kv-engine-version": "2"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.KvVersion != 2 {
			t.Errorf("expected KvVersion=2, got %d", opt.KvVersion)
		}
	})

	t.Run("invalid kv-engine-version returns error", func(t *testing.T) {
		opt := MakeOptVaultEngine()

		err := opt.UpdateFromDockerVolume("vol", map[string]string{"kv-engine-version": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid kv-engine-version")
		}
	})

	t.Run("engine-mount empty string clears the mount path", func(t *testing.T) {
		mountPath := "existing"
		opt := MakeOptVaultEngine()
		opt.MountPath = &mountPath

		if err := opt.UpdateFromDockerVolume("vol", map[string]string{"engine-mount": ""}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// empty mount path is cleaned to nil during Normalize
		if opt.MountPath != nil && *opt.MountPath != "" {
			t.Errorf("expected mount path to be cleared, got %q", *opt.MountPath)
		}
	})
}
