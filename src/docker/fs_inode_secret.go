// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"context"
	"errors"
	"os"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/anthochamp/docker-plugin-vaultfs/backend"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fsInodeSecretChild struct {
	inodeSecretField *FsInodeSecretField
	inode            *fs.Inode
}

type FsInodeSecret struct {
	fs.Inode

	secret          backend.Secret
	optDockerVolume options.OptDockerVolume

	ATime *time.Time

	lock       sync.RWMutex
	secretData backend.SecretData
	childs     map[string]fsInodeSecretChild
}

func (*FsInodeSecret) FileMode() uint32 { return fuse.S_IFDIR }

func (z *FsInodeSecret) AttrMode() uint32 { return z.optDockerVolume.MountMode }
func (z *FsInodeSecret) AttrOwner() fuse.Owner {
	return fuse.Owner{
		Uid: uint32(z.optDockerVolume.MountUId),
		Gid: uint32(z.optDockerVolume.MountGId),
	}
}

func (z *FsInodeSecret) MTime() *time.Time {
	z.lock.RLock()
	defer z.lock.RUnlock()

	if z.secretData == nil {
		return nil
	} else {
		return z.secretData.CreatedAt()
	}
}

func (z *FsInodeSecret) CTime() *time.Time { return z.MTime() }

func NewFsInodeSecret(secret backend.Secret, optDockerVolume options.OptDockerVolume) *FsInodeSecret {
	util.Tracef("NewFsInodeSecret(%+v, %+v)\n", secret, optDockerVolume)

	return &FsInodeSecret{
		secret:          secret,
		optDockerVolume: optDockerVolume,

		lock:   sync.RWMutex{},
		childs: map[string]fsInodeSecretChild{},
	}
}

func (z *FsInodeSecret) Close() {
	z.lock.Lock()
	defer z.lock.Unlock()

	z.clearCacheUnsafe()
}

func (z *FsInodeSecret) clearCacheUnsafe() {
	z.secretData = nil

	for _, v := range z.childs {
		v.inode.ForgetPersistent()
	}
	z.childs = map[string]fsInodeSecretChild{}
}

func (z *FsInodeSecret) updateCache(ctx context.Context, data backend.SecretData) syscall.Errno {
	z.lock.Lock()
	defer z.lock.Unlock()

	if z.secretData != nil && z.secretData.UniqueId() == data.UniqueId() {
		return fs.OK
	}

	z.secretData = data

	keys := data.GetKeys()

	childs := map[string]fsInodeSecretChild{}

	for _, key := range keys {
		value, ok := data.GetValue(key)
		if !ok {
			continue
		}

		child, ok := z.childs[key]
		if !ok {
			inodeSecretField := newFsInodeSecretField(z.optDockerVolume)

			inode := z.NewPersistentInode(ctx, inodeSecretField, fs.StableAttr{Mode: inodeSecretField.FileMode()})

			child = fsInodeSecretChild{
				inodeSecretField: inodeSecretField,
				inode:            inode,
			}
		}

		child.inodeSecretField.UpdateData([]byte(*value), data)
		childs[key] = child
	}

	for k, v := range z.childs {
		if !slices.Contains(keys, k) {
			v.inode.ForgetPersistent()
		}
	}

	z.childs = childs
	return fs.OK
}

func (z *FsInodeSecret) updateData(ctx context.Context, noCache bool) syscall.Errno {
	data, err := z.secret.GetData(noCache)
	if err != nil {
		z.lock.Lock()
		defer z.lock.Unlock()

		z.clearCacheUnsafe()

		util.Errorf("Unable to get secret data: %v\n", err)

		if errors.Is(err, os.ErrNotExist) {
			return syscall.ENOENT
		}

		return syscall.EIO
	}

	return z.updateCache(ctx, *data)
}

var _ = (fs.NodeReaddirer)((*FsInodeSecret)(nil))
var _ = (fs.NodeLookuper)((*FsInodeSecret)(nil))
var _ = (fs.NodeGetattrer)((*FsInodeSecret)(nil))

func (z *FsInodeSecret) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	util.Tracef("FsInodeSecret[%v].Readdir()\n", z)

	if errno := z.updateData(ctx, true); errno != fs.OK {
		return nil, errno
	}

	now := time.Now()
	z.ATime = &now

	z.lock.RLock()
	r := make([]fuse.DirEntry, 0, len(z.childs))
	for k, v := range z.childs {
		r = append(r, fuse.DirEntry{Name: k, Mode: v.inodeSecretField.FileMode()})
	}
	z.lock.RUnlock()

	//_, parent := z.Parent()
	//parent.NotifyEntry()

	return fs.NewListDirStream(r), fs.OK
}

func (z *FsInodeSecret) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	util.Tracef("FsInodeSecret[%v].Lookup(%s)\n", z, name)

	if errno := z.updateData(ctx, false); errno != fs.OK {
		return nil, errno
	}

	child, ok := z.childs[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	out.Attr.Mode = child.inodeSecretField.AttrMode()
	out.Attr.Owner = child.inodeSecretField.AttrOwner()
	out.SetTimes(child.inodeSecretField.ATime, child.inodeSecretField.MTime(), child.inodeSecretField.CTime())

	return child.inode, fs.OK
}

func (z *FsInodeSecret) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	util.Tracef("FsInodeSecret[%v].Getattr()\n", z)

	out.Mode = z.AttrMode()
	out.Owner = z.AttrOwner()
	out.SetTimes(z.ATime, z.MTime(), z.CTime())

	return fs.OK
}
