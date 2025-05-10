// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"context"
	"sync"
	"syscall"
	"time"

	"github.com/anthochamp/docker-plugin-vaultfs/backend"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type FsInodeSecretField struct {
	fs.Inode

	optDockerVolume options.OptDockerVolume

	ATime *time.Time

	lock       *sync.RWMutex
	secretData backend.SecretData
	data       []byte
}

func (*FsInodeSecretField) FileMode() uint32 { return fuse.S_IFREG }

func (z *FsInodeSecretField) AttrMode() uint32 { return z.optDockerVolume.FieldMountMode }
func (z *FsInodeSecretField) AttrOwner() fuse.Owner {
	return fuse.Owner{
		Uid: uint32(z.optDockerVolume.MountUId),
		Gid: uint32(z.optDockerVolume.MountGId),
	}
}

func (z *FsInodeSecretField) MTime() *time.Time {
	z.lock.RLock()
	defer z.lock.RUnlock()

	if z.secretData == nil {
		return nil
	} else {
		return z.secretData.CreatedAt()
	}
}

func (z *FsInodeSecretField) CTime() *time.Time { return z.MTime() }

func newFsInodeSecretField(optDockerVolume options.OptDockerVolume) *FsInodeSecretField {
	util.Tracef("newFsInodeSecretField(%v, %+v, %+v)\n", optDockerVolume)

	return &FsInodeSecretField{
		optDockerVolume: optDockerVolume,

		lock: &sync.RWMutex{},
	}
}

func (z *FsInodeSecretField) UpdateData(data []byte, secretData backend.SecretData) {
	z.lock.Lock()
	defer z.lock.Unlock()

	z.secretData = secretData
	z.data = data
}

var _ = (fs.NodeOpener)((*FsInodeSecretField)(nil))
var _ = (fs.NodeReader)((*FsInodeSecretField)(nil))
var _ = (fs.NodeGetattrer)((*FsInodeSecretField)(nil))

func (z *FsInodeSecretField) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	util.Tracef("FsInodeSecretField[%v].Open(%v)\n", z, flags)

	now := time.Now()
	z.ATime = &now

	// do not use fuse.FOPEN_DIRECT_IO, it'll prevent file from being used with mmap
	return nil, 0, fs.OK
}

func (z *FsInodeSecretField) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	util.Tracef("FsInodeSecretField[%v].Read(%+v, %v)\n", z, fh, off)

	end := int(off) + len(dest)
	if end > len(z.data) {
		end = len(z.data)
	}

	return fuse.ReadResultData(z.data[off:end]), fs.OK
}

func (z *FsInodeSecretField) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	util.Tracef("FsInodeSecretField[%v].Getattr(%+v)\n", z, fh)

	out.Mode = z.AttrMode()
	out.Owner = z.AttrOwner()
	out.SetTimes(z.ATime, z.MTime(), z.CTime())
	out.Attr.Size = uint64(len(z.data))

	return fs.OK
}
