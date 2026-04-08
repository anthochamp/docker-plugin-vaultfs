// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"bytes"
	"testing"
)

func TestCopyBuf(t *testing.T) {
	t.Run("copies all bytes to the writer", func(t *testing.T) {
		content := []byte("hello, copy buf")
		destination := &bytes.Buffer{}

		written, err := CopyBuf(destination, bytes.NewReader(content))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if written != int64(len(content)) {
			t.Errorf("expected %d bytes written, got %d", len(content), written)
		}

		if !bytes.Equal(destination.Bytes(), content) {
			t.Errorf("expected %q, got %q", content, destination.Bytes())
		}
	})

	t.Run("empty reader writes zero bytes", func(t *testing.T) {
		destination := &bytes.Buffer{}

		written, err := CopyBuf(destination, bytes.NewReader(nil))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if written != 0 {
			t.Errorf("expected 0 bytes written, got %d", written)
		}
	})

	t.Run("copies data larger than 32K buffer without corruption", func(t *testing.T) {
		large := make([]byte, 64*1024)
		for index := range large {
			large[index] = byte(index % 256)
		}

		destination := &bytes.Buffer{}

		written, err := CopyBuf(destination, bytes.NewReader(large))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if written != int64(len(large)) {
			t.Errorf("expected %d bytes written, got %d", len(large), written)
		}

		if !bytes.Equal(destination.Bytes(), large) {
			t.Error("copied content does not match source")
		}
	})
}
