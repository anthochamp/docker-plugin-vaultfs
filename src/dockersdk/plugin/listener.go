// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type CleanupFn func() error

type Listener struct {
	cleanupFn CleanupFn
	listener  net.Listener

	server    *http.Server
	closeLock *sync.Mutex
	closed    bool
}

// NewListenerUnixSocket("/run/docker/plugins/plugin.sock", 0, -1, 0660)
func NewListenerUnixSocket(socketPath string, uid int, gid int, mode uint32) (*Listener, error) {
	// force unix socket access mode to X000 on creation
	initialUmask := syscall.Umask(0777)

	listener, err := net.Listen("unix", socketPath)

	syscall.Umask(initialUmask)

	if err != nil {
		return nil, err
	}

	if err := os.Chown(socketPath, uid, gid); err != nil {
		listener.Close()
		return nil, fmt.Errorf("update unix socket owner to %d:%d: %w", uid, gid, err)
	}

	if err := os.Chmod(socketPath, fs.FileMode(mode)); err != nil {
		listener.Close()
		return nil, fmt.Errorf("update unix socket access modes: %w", err)
	}

	return &Listener{
		listener: listener,
		cleanupFn: func() error {
			err := os.Remove(socketPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			return nil
		},

		closeLock: &sync.Mutex{},
	}, nil
}

// TCP Socket requires the creation of a spec file in the docker configuration directory :
//   - /etc/docker/plugins/pluginname.spec
//   - \%ProgramData%\docker\plugins\pluginname.spec
//
// which contains the plugin IP and TCP port to connect to eg. :
// tcp://[ip]:[port]
func NewListenerTcpSocket(bindAddress string, bindPort uint16, tlsConfig *tls.Config) (*Listener, error) {
	var address string

	if bindAddress == "" {
		address = "0.0.0.0"
	} else {
		address = bindAddress
	}

	if bindPort == 0 {
		return nil, fmt.Errorf("bind port cannot be zero")
	}

	address += ":" + strconv.Itoa(int(bindPort))

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("create TCP socket %s: %w", address, err)
	}

	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"http/1.1"}
		listener = tls.NewListener(listener, tlsConfig)
	}

	return &Listener{
		listener: listener,

		closeLock: &sync.Mutex{},
	}, nil
}

// Windows Named Pipe requires the creation of a .spec file in the docker configuration directory :
//   - \%ProgramData%\docker\plugins\pluginname.spec
//
// which contains the plugin named pipe to connect to eg. :
// npipe://[pipename]
func NewListenerWindowsNamedPipe() (*Listener, error) {
	return nil, errors.New("not implemented")
}

func (z *Listener) Serve(plugin Plugin) error {
	util.Tracef("dockerSdkPlugin.Listener.Serve()\n")

	if z.server != nil {
		return nil
	}

	if z.closed {
		return errors.New("listener has been closed")
	}

	z.server = &http.Server{
		Handler: plugin.serveMux,
	}

	err := z.server.Serve(z.listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (z *Listener) Close() error {
	util.Tracef("dockerSdkPlugin.Listener.Close()\n")

	if !z.closeLock.TryLock() {
		return nil
	}
	defer z.closeLock.Unlock()

	if z.closed {
		return nil
	}
	z.closed = true

	if z.server != nil {
		// http.server.Close *should* close the listener
		if err := z.server.Close(); err != nil {
			return err
		}

		z.server = nil
	} else {
		if err := z.listener.Close(); err != nil {
			return err
		}
	}

	defer func() {
		if z.cleanupFn != nil {
			if err := z.cleanupFn(); err != nil {
				util.Errorf("Unable to cleanup listener: %v\n", err)
			}
		}
	}()

	return nil
}
