// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Fs struct {
	FsConfig

	InodeRoot *FsInodeRoot

	lock       *sync.Mutex
	fuseServer *fuse.Server
}

type FsConfig struct {
	MountFuseName string
	MountDir      string
	MountDirUId   uint16
	MountDirGId   uint16
}

func newFs(config FsConfig) *Fs {
	return &Fs{
		FsConfig: config,

		InodeRoot: newFsInodeRoot(),

		lock: &sync.Mutex{},
	}
}

func (z *Fs) Close() {
	z.InodeRoot.Close()
	z.InodeRoot = nil
}

func (z *Fs) Mount() error {
	util.Tracef("Fs[%v].Mount()\n", z)

	z.lock.Lock()
	defer z.lock.Unlock()

	if z.fuseServer != nil {
		return nil
	}

	// Ensure filesystem is not already mounted
	cmd := exec.Command("fusermount", "-u", z.MountDir)
	if err := cmd.Run(); err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			return fmt.Errorf("assert FS directory is unmounted: %w", err)
		}
	}

	// Create filesystem mount directory
	const fsMountDirPerm = 0o550
	if err := os.MkdirAll(z.MountDir, fsMountDirPerm); err != nil {
		return fmt.Errorf("create FS mount directory: %w", err)
	}

	// Mount FS
	fuseServer, err := fs.Mount(z.MountDir, z.InodeRoot, &fs.Options{
		UID: uint32(z.MountDirUId),
		GID: uint32(z.MountDirGId),
		MountOptions: fuse.MountOptions{
			// Allows processes not on UID/GID to access the mounted filesystem
			// Rationals: Docker's containers can runs on any UID/GID but must
			// still be able to access mounted volumes via the FUSE filesystem.
			AllowOther: true,

			Name:  z.MountFuseName,
			Debug: util.DebugMode,
		},
	})
	if err != nil {
		return err
	}

	z.fuseServer = fuseServer
	return nil
}

func (z *Fs) Unmount() error {
	util.Tracef("Fs[%v].Unmount()\n", z)

	z.lock.Lock()
	defer z.lock.Unlock()

	if z.fuseServer == nil {
		return nil
	}

	// Note: Will block if any dir/file lock is hold (= if any docker container is started)
	err := z.fuseServer.Unmount()
	// "Invalid argument" is when the file system is already unmounted
	if err != nil && !strings.Contains(err.Error(), "Invalid argument") {
		return err
	}

	z.fuseServer = nil
	return nil
}

func (z *Fs) WaitUnmount() error {
	var fuseServer *fuse.Server

	z.lock.Lock()
	fuseServer = z.fuseServer
	z.lock.Unlock()

	if fuseServer == nil {
		return errors.New("FS is not mounted")
	}

	fuseServer.Wait()

	return nil
}
