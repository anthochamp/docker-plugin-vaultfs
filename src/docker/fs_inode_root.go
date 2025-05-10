// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"context"
	"syscall"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/hanwen/go-fuse/v2/fs"
)

type FsInodeRoot struct {
	fs.Inode
}

func newFsInodeRoot() *FsInodeRoot {
	return &FsInodeRoot{}
}

func (z *FsInodeRoot) Close() {
	children := z.Children()

	for name, child := range children {
		z.RmChild(name)
		z.NotifyDelete(name, child)
		child.ForgetPersistent()
	}
}

func (z *FsInodeRoot) addInodeSecret(name string, inodeSecret *FsInodeSecret) error {
	util.Tracef("FsInodeRoot[%v].addInodeSecret(%s, %+v)\n", z, name, inodeSecret)

	child := z.NewPersistentInode(context.Background(), inodeSecret, fs.StableAttr{Mode: inodeSecret.FileMode()})

	ok := z.AddChild(name, child, false)
	if !ok {
		return syscall.EEXIST
	}

	return nil
}

func (z *FsInodeRoot) removeInodeSecret(name string) (ok bool) {
	util.Tracef("FsInodeRoot[%v].removeInodeSecret(%s)\n", z, name)

	child := z.GetChild(name)
	if child == nil {
		return false
	}

	success, _ := z.RmChild(name)
	if !success {
		return false
	}

	z.NotifyDelete(name, child)

	child.ForgetPersistent()
	return true
}
