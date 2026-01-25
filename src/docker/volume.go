// SPDX-FileCopyrightText: © 2024 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"fmt"
	"path"
	"sync"

	"github.com/anthochamp/docker-plugin-vaultfs/backend"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

// Volume represents a managed secret volume in the plugin.
type Volume struct {
	VolumeConfig

	lock            *sync.Mutex
	mountRequestIds map[string]bool
	secret          backend.Secret
	fsInodeSecret   *FsInodeSecret
	mountPath       *string
}

// MountPath returns the mount path of the volume, or an empty string if not set.
func (z *Volume) MountPath() string {
	if z.mountPath == nil {
		return ""
	} else {
		return *z.mountPath
	}
}

// VolumeConfig holds configuration for a Volume.
type VolumeConfig struct {
	Name      string            `json:","`
	OptDocker options.OptDocker `json:","`
}

// newVolume creates a new Volume with the given configuration.
func newVolume(config VolumeConfig) (*Volume, error) {
	return &Volume{
		VolumeConfig: config,

		lock:            &sync.Mutex{},
		mountRequestIds: map[string]bool{},
	}, nil
}

// mount mounts the volume, creating secrets and inodes as needed.
func (z *Volume) mount(fs *Fs, requestId string) error {
	util.Tracef("Volume[%s].mount(%s)\n", z.Name, requestId)

	z.lock.Lock()
	defer z.lock.Unlock()

	if len(z.mountRequestIds) == 0 {
		secret, err := newSecret(SecretConfig{
			OptSecret: z.OptDocker.Secret,
		})
		if err != nil {
			return fmt.Errorf("create secret: %w", err)
		}

		fsInodeSecret := NewFsInodeSecret(*secret, z.OptDocker.DockerVolume)

		if err := fs.InodeRoot.addInodeSecret(z.Name, fsInodeSecret); err != nil {
			return fmt.Errorf("add secret inode to root inode: %w", err)
		}

		z.fsInodeSecret = fsInodeSecret
		z.secret = *secret
		mountPath := path.Join(fs.MountDir, z.Name)
		z.mountPath = &mountPath
	}

	z.mountRequestIds[requestId] = true
	return nil
}

// unmount unmounts the volume for the given request ID, cleaning up resources if needed.
func (z *Volume) unmount(fs *Fs, requestId string) error {
	util.Tracef("Volume[%s].unmount(%s)\n", z.Name, requestId)

	z.lock.Lock()
	defer z.lock.Unlock()

	if _, ok := z.mountRequestIds[requestId]; !ok {
		return fmt.Errorf("unable to find mount request id %s", requestId)
	}

	if len(z.mountRequestIds) == 1 {
		if ok := fs.InodeRoot.removeInodeSecret(z.Name); !ok {
			return fmt.Errorf("remove secret inode from root inode")
		}

		delete(z.mountRequestIds, requestId)

		z.mountPath = nil
		z.secret.Close()
		z.secret = nil
		z.fsInodeSecret.Close()
		z.fsInodeSecret = nil
	}

	return nil
}

// forceUnmount forcibly unmounts the volume, cleaning up all resources.
func (z *Volume) forceUnmount(fs *Fs) error {
	util.Tracef("Volume[%s].forceUnmount()\n", z.Name)

	z.lock.Lock()
	defer z.lock.Unlock()

	if len(z.mountRequestIds) != 0 {
		if ok := fs.InodeRoot.removeInodeSecret(z.Name); !ok {
			return fmt.Errorf("remove secret inode from root inode")
		}

		z.mountRequestIds = map[string]bool{}

		z.mountPath = nil
		z.secret.Close()
		z.secret = nil
		z.fsInodeSecret.Close()
		z.fsInodeSecret = nil
	}

	return nil
}
