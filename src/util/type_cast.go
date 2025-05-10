// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

func MapStringStringFromMapStringInterface(m map[string]interface{}) map[string]string {
	r := map[string]string{}
	for k, v := range m {
		r[k] = v.(string)
	}
	return r
}
