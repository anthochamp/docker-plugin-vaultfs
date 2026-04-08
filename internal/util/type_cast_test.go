// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"testing"
)

func TestMapStringStringFromMapStringInterface(t *testing.T) {
	t.Run("empty map returns empty result", func(t *testing.T) {
		result := MapStringStringFromMapStringInterface(map[string]interface{}{})

		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("converts each value to its string representation", func(t *testing.T) {
		input := map[string]interface{}{
			"alpha": "valueA",
			"beta":  "valueB",
		}

		result := MapStringStringFromMapStringInterface(input)

		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}

		if result["alpha"] != "valueA" {
			t.Errorf("expected alpha=%q, got %q", "valueA", result["alpha"])
		}

		if result["beta"] != "valueB" {
			t.Errorf("expected beta=%q, got %q", "valueB", result["beta"])
		}
	})
}
