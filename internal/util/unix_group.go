// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"errors"
	"fmt"
	os_user "os/user"
	"strconv"
)

func UserIdFromUser(user string) (uint16, error) {
	uid_, err := strconv.Atoi(user)
	if err != nil || uid_ < 0 {
		user, err := os_user.Lookup(user)
		if err != nil {
			return 0, fmt.Errorf("unable to find id of user %s: %w", user, err)
		}

		uid_, _ = strconv.Atoi(user.Uid)
		if uid_ < 0 {
			return 0, errors.New(fmt.Sprintf("Unable to convert user id %s", user.Uid))
		}
	}

	return uint16(uid_), nil
}

func UserGroupIdFromUserGroup(userGroup string) (uint16, error) {
	gid_, err := strconv.Atoi(userGroup)
	if err != nil || gid_ < 0 {
		group, err := os_user.LookupGroup(userGroup)
		if err != nil {
			return 0, fmt.Errorf("unable to find id of group %s: %w", userGroup, err)
		}

		gid_, _ = strconv.Atoi(group.Gid)
		if gid_ < 0 {
			return 0, errors.New(fmt.Sprintf("Unable to convert group id %s", group.Gid))
		}
	}

	return uint16(gid_), nil
}
