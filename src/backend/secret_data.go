// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package backend

import (
	"time"
)

type SecretData interface {
	UniqueId() string

	CreatedAt() *time.Time

	GetKeys() []string
	GetValue(key string) (*string, bool)
}
