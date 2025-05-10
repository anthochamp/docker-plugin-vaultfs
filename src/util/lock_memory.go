// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package util

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

func LockMemory() error {
	err := unix.Mlockall(unix.MCL_FUTURE | unix.MCL_CURRENT)

	switch err {
	case nil:
		return nil
	case unix.ENOMEM:
		err = errors.New("Process has a nonzero RLIMIT_MEMLOCK soft resource limit and CAP_IPC_LOCK is not set (see ENOMEM err at https://linux.die.net/man/2/mlockall)")
	case unix.EPERM:
		err = errors.New("Process requires CAP_IPC_LOCK to be set (see EPERM err at https://linux.die.net/man/2/mlockall)")
	}

	return fmt.Errorf("lock memory: %w", err)
}
