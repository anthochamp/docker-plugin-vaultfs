// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package docker

import (
	"crypto/tls"
	"fmt"
	"sync"

	dockerSdkPlugin "github.com/anthochamp/docker-plugin-vaultfs/dockersdk/plugin"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type Plugin struct {
	volumeDriver *VolumeDriver

	listener *dockerSdkPlugin.Listener
	plugin   dockerSdkPlugin.Plugin

	cleanUpLock *sync.Mutex
	doneChan    chan bool
}

type PluginConfig struct {
	TcpBindAddr    *string
	TcpBindPort    *uint16
	TcpTlsConfig   *tls.Config
	UnixSocketPath *string
	UnixSocketUId  uint16
	UnixSocketGId  uint16
	UnixSocketMode uint32

	VolumeDriverDisabled      bool
	VolumeDriverFsConfig      FsConfig
	VolumeDriverGlobalScope   bool
	VolumeDriverStateFilePath string

	SecretProviderDisabled bool

	DefaultOptDocker options.OptDocker
}

func NewPlugin(config PluginConfig) (*Plugin, error) {
	var volumeDriver *VolumeDriver
	var err error

	plugin := dockerSdkPlugin.MakePlugin()

	if !config.VolumeDriverDisabled {
		volumeDriver, err = NewVolumeDriver(
			VolumeDriverConfig{
				FsConfig: config.VolumeDriverFsConfig,

				GlobalScope:   config.VolumeDriverGlobalScope,
				StateFilePath: config.VolumeDriverStateFilePath,

				DefaultOptDocker: config.DefaultOptDocker,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("create volume driver: %w", err)
		}

		dockerSdkPlugin.RegisterVolumeDriver(volumeDriver, plugin)
	}

	if !config.SecretProviderDisabled {
		secretProvider, err := NewSecretProvider(
			SecretProviderConfig{
				DefaultOptDocker: config.DefaultOptDocker,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("create secret provider: %w", err)
		}

		dockerSdkPlugin.RegisterSecretProvider(secretProvider, plugin)
	}

	var listener *dockerSdkPlugin.Listener

	if config.TcpBindAddr != nil && config.TcpBindPort != nil {
		listener, err = dockerSdkPlugin.NewListenerTcpSocket(*config.TcpBindAddr, *config.TcpBindPort, config.TcpTlsConfig)
		if err != nil {
			return nil, fmt.Errorf("create tcp listener: %w", err)
		}
	} else if config.UnixSocketPath != nil {
		listener, err = dockerSdkPlugin.NewListenerUnixSocket(*config.UnixSocketPath, int(config.UnixSocketUId), int(config.UnixSocketGId), config.UnixSocketMode)
		if err != nil {
			return nil, fmt.Errorf("create unix socket listener: %w", err)
		}
	}

	return &Plugin{
		volumeDriver: volumeDriver,

		plugin:   plugin,
		listener: listener,

		cleanUpLock: &sync.Mutex{},
		doneChan:    make(chan bool, 1),
	}, nil
}

func (z *Plugin) Initialize() error {
	util.Tracef("Plugin.Initialize()\n")

	if z.volumeDriver != nil {
		if err := z.volumeDriver.Initialize(); err != nil {
			return fmt.Errorf("initialize volume driver: %w", err)
		}

		go func() {
			<-z.volumeDriver.DoneChan()

			util.Printf("Volume driver closed\n")

			z.CleanUp()
		}()
	}

	go func() {
		err := z.listener.Serve(z.plugin)

		if err == nil {
			util.Printf("Plugin serve completed\n")
		} else {
			util.Errorf("Plugin serve completed with error: %v", err)
		}

		z.CleanUp()
	}()

	return nil
}

func (z *Plugin) CleanUp() {
	util.Tracef("Plugin.CleanUp()\n")

	if !z.cleanUpLock.TryLock() {
		return
	}
	defer z.cleanUpLock.Unlock()

	if z.listener != nil {
		if err := z.listener.Close(); err != nil {
			util.Errorf("close Docker listener: %v", err)
		} else {
			z.listener = nil
		}
	}

	if z.volumeDriver != nil {
		z.volumeDriver.CleanUp()
		z.volumeDriver = nil
	}

	z.doneChan <- true
}

func (z Plugin) DoneChan() chan bool {
	return z.doneChan
}
