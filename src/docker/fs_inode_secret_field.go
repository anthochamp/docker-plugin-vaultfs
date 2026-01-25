// SPDX-FileCopyrightText: © 2024 - 2026 Anthony Champagne <dev@anthonychampagne.fr>
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

type FsInodeSecretFieldFileHandle struct {
	data []byte
}

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
var _ = (fs.NodeGetattrer)((*FsInodeSecretField)(nil))
var _ = (fs.FileReader)((*FsInodeSecretFieldFileHandle)(nil))

func (z *FsInodeSecretField) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	util.Tracef("FsInodeSecretField[%v].Open(%v)\n", z, flags)

	now := time.Now()
	z.ATime = &now

	z.lock.RLock()
	dataCopy := make([]byte, len(z.data))
	copy(dataCopy, z.data)
	z.lock.RUnlock()

	return &FsInodeSecretFieldFileHandle{data: dataCopy}, 0, fs.OK
}

func (z *FsInodeSecretFieldFileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	util.Tracef("FsInodeSecretFieldFileHandle.Read(%v)\n", off)

	end := int(off) + len(dest)
	if end > len(z.data) {
		end = len(z.data)
	}

	return fuse.ReadResultData(z.data[off:end]), fs.OK
}

func (z *FsInodeSecretField) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	util.Tracef("FsInodeSecretField[%v].Getattr(%+v)\n", z, fh)

	z.lock.RLock()
	defer z.lock.RUnlock()

	out.Mode = z.AttrMode()
	out.Owner = z.AttrOwner()
	out.SetTimes(z.ATime, z.MTime(), z.CTime())
	out.Attr.Size = uint64(len(z.data))
	out.SetTimeout(1 * time.Second)

	return fs.OK
}
