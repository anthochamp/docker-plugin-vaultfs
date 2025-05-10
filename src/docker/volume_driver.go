// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	dockerSdkPlugin "github.com/anthochamp/docker-plugin-vaultfs/dockersdk/plugin"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type VolumeDriver struct {
	VolumeDriverConfig

	fs *Fs

	cleanUpLock *sync.Mutex

	doneChan chan bool

	volumesLock *sync.RWMutex
	volumes     map[string]*Volume
}

type VolumeDriverConfig struct {
	FsConfig FsConfig

	GlobalScope   bool
	StateFilePath string

	DefaultOptDocker options.OptDocker
}

func NewVolumeDriver(config VolumeDriverConfig) (*VolumeDriver, error) {
	return &VolumeDriver{
		VolumeDriverConfig: config,

		fs: newFs(config.FsConfig),

		cleanUpLock: &sync.Mutex{},

		doneChan: make(chan bool, 1),

		volumesLock: &sync.RWMutex{},
		volumes:     map[string]*Volume{},
	}, nil
}

func (z *VolumeDriver) Initialize() error {
	util.Tracef("VolumeDriver.Initialize()\n")

	if err := z.restoreVolumes(); err != nil {
		if errb := z.fs.Unmount(); errb != nil {
			util.Errorf("Unable to unmount volume FS: %v\n", errb)
		}

		return fmt.Errorf("restore volumes: %w", err)
	}

	if err := z.fs.Mount(); err != nil {
		return fmt.Errorf("mount volume FS: %w", err)
	}

	go func() {
		z.fs.WaitUnmount()

		util.Printf("Volume FS unmounted\n")

		z.CleanUp()
	}()

	return nil
}

func (z *VolumeDriver) CleanUp() {
	util.Tracef("VolumeDriver.CleanUp()\n")

	if !z.cleanUpLock.TryLock() {
		return
	}
	defer z.cleanUpLock.Unlock()

	if err := z.fs.Unmount(); err != nil {
		util.Errorf("unmount volume FS: %v", err)
	}

	z.volumesLock.Lock()
	for _, v := range z.volumes {
		v.forceUnmount(z.fs)
	}
	z.volumes = map[string]*Volume{}
	z.volumesLock.Unlock()

	z.doneChan <- true
}

func (z VolumeDriver) DoneChan() chan bool {
	return z.doneChan
}

func (z VolumeDriver) backupVolumes() error {
	util.Tracef("VolumeDriver.backupVolumes()\n")

	if err := os.MkdirAll(path.Dir(z.StateFilePath), 0770); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	var volumesBackup []VolumeConfig

	z.volumesLock.RLock()
	volumesBackup = make([]VolumeConfig, 0, len(z.volumes))
	for _, v := range z.volumes {
		volumesBackup = append(volumesBackup, v.VolumeConfig)
	}
	z.volumesLock.RUnlock()

	fileData, err := json.Marshal(volumesBackup)
	if err != nil {
		return fmt.Errorf("serialize volume backup data: %w", err)
	}

	return os.WriteFile(z.StateFilePath, fileData, 0600)
}

func (z VolumeDriver) restoreVolumes() error {
	util.Tracef("VolumeDriver.restoreVolumes()\n")

	content, err := os.ReadFile(z.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("read volume backup: %w", err)
	}

	var volumesData []VolumeConfig
	err = json.Unmarshal(content, &volumesData)
	if err != nil {
		return fmt.Errorf("unserialize volume backup content: %w", err)
	}

	for _, volumeConfig := range volumesData {
		volume, err := newVolume(volumeConfig)
		if err != nil {
			return fmt.Errorf("create volume %s: %w", volumeConfig.Name, err)
		}

		z.volumes[volumeConfig.Name] = volume
	}

	return nil
}

/***/

func (z VolumeDriver) Create(r dockerSdkPlugin.VolumeDriverCreateRequest) error {
	util.Tracef("VolumeDriver.Create(%+v)\n", r)

	err := func() error {
		z.volumesLock.Lock()
		defer z.volumesLock.Unlock()

		if _, ok := z.volumes[r.Name]; ok {
			return fmt.Errorf("volume %s already exists", r.Name)
		}

		optDocker, err := options.NewOptDockerFromDockerVolume(r.Name, r.Options, &z.DefaultOptDocker)
		if err != nil {
			return fmt.Errorf("compose secrets options: %w", err)
		}

		v, err := newVolume(VolumeConfig{
			Name:      r.Name,
			OptDocker: *optDocker,
		})
		if err != nil {
			return fmt.Errorf("create volume %s: %w", r.Name, err)
		}

		z.volumes[r.Name] = v
		return nil
	}()
	if err != nil {
		return err
	}

	if err = z.backupVolumes(); err != nil {
		util.Errorf("Unable to backup volumes: %v\n", err)
	}

	return nil
}

func (z VolumeDriver) List() (*dockerSdkPlugin.VolumeDriverListResponse, error) {
	util.Tracef("VolumeDriver.List()\n")

	z.volumesLock.RLock()
	defer z.volumesLock.RUnlock()

	r := make([]*dockerSdkPlugin.Volume, 0, len(z.volumes))

	for _, v := range z.volumes {
		r = append(r, &dockerSdkPlugin.Volume{Name: v.Name, Mountpoint: v.MountPath()})
	}

	return &dockerSdkPlugin.VolumeDriverListResponse{Volumes: r}, nil
}

func (z VolumeDriver) Get(r dockerSdkPlugin.VolumeDriverGetRequest) (*dockerSdkPlugin.VolumeDriverGetResponse, error) {
	util.Tracef("VolumeDriver.Get(%+v)\n", r)

	z.volumesLock.RLock()
	defer z.volumesLock.RUnlock()

	v, ok := z.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find volume %s", r.Name)
	}

	return &dockerSdkPlugin.VolumeDriverGetResponse{Volume: &dockerSdkPlugin.Volume{Name: v.Name, Mountpoint: v.MountPath()}}, nil
}

func (z VolumeDriver) Remove(r dockerSdkPlugin.VolumeDriverRemoveRequest) error {
	util.Tracef("VolumeDriver.Remove(%+v)\n", r)

	err := func() error {
		z.volumesLock.Lock()
		defer z.volumesLock.Unlock()

		v, ok := z.volumes[r.Name]
		if !ok {
			return fmt.Errorf("unable to find volume %s", r.Name)
		}

		if err := v.forceUnmount(z.fs); err != nil {
			return fmt.Errorf("force volume unmount %s: %w", r.Name, err)
		}

		delete(z.volumes, r.Name)
		return nil
	}()
	if err != nil {
		return err
	}

	if err := z.backupVolumes(); err != nil {
		util.Errorf("Unable to backup volumes: %v\n", err)
	}

	return nil
}

func (z VolumeDriver) Path(r dockerSdkPlugin.VolumeDriverPathRequest) (*dockerSdkPlugin.VolumeDriverPathResponse, error) {
	util.Tracef("VolumeDriver.Path(%+v)\n", r)

	z.volumesLock.RLock()
	defer z.volumesLock.RUnlock()

	v, ok := z.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find volume %s", r.Name)
	}

	return &dockerSdkPlugin.VolumeDriverPathResponse{Mountpoint: v.MountPath()}, nil
}

func (z VolumeDriver) Mount(r dockerSdkPlugin.VolumeDriverMountRequest) (*dockerSdkPlugin.VolumeDriverMountResponse, error) {
	util.Tracef("VolumeDriver.Mount(%+v)\n", r)

	z.volumesLock.RLock()
	defer z.volumesLock.RUnlock()

	v, ok := z.volumes[r.Name]
	if !ok {
		return nil, fmt.Errorf("unable to find volume %s", r.Name)
	}

	if err := v.mount(z.fs, r.ID); err != nil {
		return nil, fmt.Errorf("mount volume %s: %w", r.Name, err)
	}

	return &dockerSdkPlugin.VolumeDriverMountResponse{Mountpoint: v.MountPath()}, nil
}

func (z VolumeDriver) Unmount(r dockerSdkPlugin.VolumeDriverUnmountRequest) error {
	util.Tracef("VolumeDriver.Unmount(%+v)\n", r)

	z.volumesLock.RLock()
	defer z.volumesLock.RUnlock()

	v, ok := z.volumes[r.Name]
	if !ok {
		return fmt.Errorf("unable to find volume %s", r.Name)
	}

	if err := v.unmount(z.fs, r.ID); err != nil {
		return fmt.Errorf("unmount volume %s: %w", r.Name, err)
	}

	return nil
}

func (z VolumeDriver) Capabilities() (*dockerSdkPlugin.VolumeDriverCapabilitiesResponse, error) {
	util.Tracef("VolumeDriver.Capabilities()\n")

	var scope string
	if z.GlobalScope {
		scope = "global"
	} else {
		scope = "local"
	}

	return &dockerSdkPlugin.VolumeDriverCapabilitiesResponse{
		Capabilities: dockerSdkPlugin.VolumeDriverCapability{
			Scope: scope,
		},
	}, nil
}
