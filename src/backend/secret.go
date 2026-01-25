// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package backend

// Secret represents a generic secret with lifecycle and data retrieval methods.
type Secret interface {
	Close()

	GetData(noCache bool) (*SecretData, error)
}
