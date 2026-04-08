// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"testing"
)

func TestOptDockerVolumeDefaults(t *testing.T) {
	t.Run("default mount mode is 0550", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if opt.MountMode != 0o550 {
			t.Errorf("expected MountMode=0550 (octal), got %04o", opt.MountMode)
		}
	})

	t.Run("default field mount mode is 0440", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if opt.FieldMountMode != 0o440 {
			t.Errorf("expected FieldMountMode=0440 (octal), got %04o", opt.FieldMountMode)
		}
	})
}

func TestOptDockerVolumeUpdate(t *testing.T) {
	t.Run("mount-uid option is parsed as integer", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if err := opt.Update("vol", map[string]string{"mount-uid": "1000"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountUId != 1000 {
			t.Errorf("expected MountUId=1000, got %d", opt.MountUId)
		}
	})

	t.Run("mount-gid option is parsed as integer", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if err := opt.Update("vol", map[string]string{"mount-gid": "2000"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountGId != 2000 {
			t.Errorf("expected MountGId=2000, got %d", opt.MountGId)
		}
	})

	t.Run("mount-mode option is parsed as integer", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if err := opt.Update("vol", map[string]string{"mount-mode": "493"}); err != nil { // 493 = 0755
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountMode != 493 {
			t.Errorf("expected MountMode=493, got %d", opt.MountMode)
		}
	})

	t.Run("field-mount-mode option is parsed as integer", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if err := opt.Update("vol", map[string]string{"field-mount-mode": "256"}); err != nil { // 256 = 0400
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.FieldMountMode != 256 {
			t.Errorf("expected FieldMountMode=256, got %d", opt.FieldMountMode)
		}
	})

	t.Run("invalid mount-uid returns error", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		err := opt.Update("vol", map[string]string{"mount-uid": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid mount-uid")
		}
	})

	t.Run("invalid mount-gid returns error", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		err := opt.Update("vol", map[string]string{"mount-gid": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid mount-gid")
		}
	})

	t.Run("invalid mount-mode returns error", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		err := opt.Update("vol", map[string]string{"mount-mode": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid mount-mode")
		}
	})

	t.Run("invalid field-mount-mode returns error", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		err := opt.Update("vol", map[string]string{"field-mount-mode": "not-a-number"})

		if err == nil {
			t.Error("expected error for invalid field-mount-mode")
		}
	})

	t.Run("absent options leave defaults unchanged", func(t *testing.T) {
		opt := MakeOptDockerVolume()

		if err := opt.Update("vol", map[string]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if opt.MountUId != 0 {
			t.Errorf("expected MountUId=0, got %d", opt.MountUId)
		}

		if opt.MountGId != 0 {
			t.Errorf("expected MountGId=0, got %d", opt.MountGId)
		}

		if opt.MountMode != 0o550 {
			t.Errorf("expected MountMode=0550, got %04o", opt.MountMode)
		}

		if opt.FieldMountMode != 0o440 {
			t.Errorf("expected FieldMountMode=0440, got %04o", opt.FieldMountMode)
		}
	})
}
