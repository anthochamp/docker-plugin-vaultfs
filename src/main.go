// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"syscall"

	"github.com/anthochamp/docker-plugin-vaultfs/constants"
	"github.com/anthochamp/docker-plugin-vaultfs/docker"
	dockerSdkPlugin "github.com/anthochamp/docker-plugin-vaultfs/dockersdk/plugin"
	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	cli "github.com/urfave/cli/v2"
)

var (
	appVersion string
	commitHash string
	buildDate  string
)

func main() {
	currentUser, _ := user.Current()
	currentGroup, _ := user.LookupGroupId(strconv.Itoa(os.Getgid()))

	defaultOptDocker := options.MakeOptDocker()

	app := &cli.App{
		Name:    constants.AppName,
		Version: fmt.Sprintf("%s, build %s+%s (%s-%s-%s)", appVersion, commitHash, buildDate, runtime.Compiler, runtime.GOOS, runtime.GOARCH),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				EnvVars:     []string{constants.EnvVarsPrefix + "DEBUG"},
				Value:       false,
				Usage:       "Debug mode (WARNING: it will leak sensitive data in logs)",
				Destination: &util.DebugMode,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				EnvVars:     []string{constants.EnvVarsPrefix + "VERBOSE"},
				Value:       false,
				Usage:       "Verbose mode",
				Destination: &util.Verbose,
			},
			&cli.BoolFlag{
				Name:    "disable-mlock",
				EnvVars: []string{constants.EnvVarsPrefix + "DISABLE_MLOCK"},
				Value:   false,
				Usage:   "Disable memory locking (NOT RECOMMENDED)",
			},
			&cli.StringFlag{
				Category:    "Vault Client Options",
				Name:        "vault-url",
				EnvVars:     []string{constants.EnvVarsPrefix + "VAULT_URL", "VAULT_ADDR"},
				Usage:       "URL of the Vault Server",
				Destination: &defaultOptDocker.Secret.Vault.ClientHttp.Address,
				Required:    true,
			},
			&cli.BoolFlag{
				Category:    "Vault Client Options",
				Name:        "vault-disable-redirects",
				EnvVars:     []string{constants.EnvVarsPrefix + "VAULT_DISABLE_REDIRECTS"},
				Value:       defaultOptDocker.Secret.Vault.ClientHttp.DisableRedirects,
				Usage:       "Disable Vault HTTP redirects",
				Destination: &defaultOptDocker.Secret.Vault.ClientHttp.DisableRedirects,
			},
			&cli.BoolFlag{
				Category:    "Vault Client Options",
				Name:        "vault-tls-skip-verify",
				EnvVars:     []string{constants.EnvVarsPrefix + "VAULT_TLS_SKIP_VERIFY", "VAULT_SKIP_VERIFY"},
				Value:       defaultOptDocker.Secret.Vault.ClientHttp.Tls.Insecure,
				Usage:       "Skip verification of Vault server TLS certificate",
				Destination: &defaultOptDocker.Secret.Vault.ClientHttp.Tls.Insecure,
			},

			&cli.StringFlag{
				Category:    "Vault Client Options",
				Name:        "auth-method",
				EnvVars:     []string{constants.EnvVarsPrefix + "AUTH_METHOD"},
				Value:       defaultOptDocker.Secret.Vault.VaultAuth.Method,
				Usage:       "Default Auth method (AppRole, Cert, Token, Userpass)",
				Destination: &defaultOptDocker.Secret.Vault.VaultAuth.Method,
			},
			&cli.StringFlag{
				Category:    "Vault Client Options",
				Name:        "auth-mount",
				EnvVars:     []string{constants.EnvVarsPrefix + "AUTH_MOUNT"},
				Value:       "",
				DefaultText: defaultOptDocker.Secret.Vault.VaultAuth.EffectiveMountPath(),
				Usage:       "Default Auth engine mount path",
			},

			&cli.StringFlag{
				Category:    "Vault Secrets",
				Name:        "engine-type",
				EnvVars:     []string{constants.EnvVarsPrefix + "ENGINE_TYPE"},
				Value:       defaultOptDocker.Secret.Vault.VaultEngine.Type,
				Usage:       "Default Vault Secrets engine type (KV, DB or PKI)",
				Destination: &defaultOptDocker.Secret.Vault.VaultEngine.Type,
			},
			&cli.StringFlag{
				Category:    "Vault Secrets",
				Name:        "engine-mount",
				EnvVars:     []string{constants.EnvVarsPrefix + "ENGINE_MOUNT"},
				Value:       "",
				DefaultText: defaultOptDocker.Secret.Vault.VaultEngine.EffectiveMountPath(),
				Usage:       "Default Vault Secrets engine mount path",
			},
			&cli.IntFlag{
				Category:    "Vault Secrets",
				Name:        "kv-engine-version",
				EnvVars:     []string{constants.EnvVarsPrefix + "KV_ENGINE_VERSION"},
				Value:       defaultOptDocker.Secret.Vault.VaultEngine.KvVersion,
				Usage:       "Default Vault Secrets K/V engine version (1 or 2)",
				Destination: &defaultOptDocker.Secret.Vault.VaultEngine.KvVersion,
			},

			&cli.StringFlag{
				Category: "Docker Plugin",
				Name:     "plugin-tcp-bind-addr",
				EnvVars:  []string{constants.EnvVarsPrefix + "PLUGIN_TCP_BIND_ADDR"},
				Value:    "0.0.0.0",
				Usage:    "Docker Plugin TCP bind address (IPv4 or IPv6)",
			},
			&cli.UintFlag{
				Category:    "Docker Plugin",
				Name:        "plugin-tcp-bind-port",
				EnvVars:     []string{constants.EnvVarsPrefix + "PLUGIN_TCP_BIND_PORT"},
				Value:       0,
				DefaultText: "<undefined>",
				Usage:       "Docker Plugin TCP bind port",
			},
			&cli.StringFlag{
				Category: "Docker Plugin",
				Name:     "plugin-socket-path",
				EnvVars:  []string{constants.EnvVarsPrefix + "PLUGIN_SOCKET_PATH"},
				Value:    "/run/docker/plugins/" + constants.DockerPluginId + ".sock",
				Usage:    "Docker Plugin Unix socket path",
			},
			&cli.StringFlag{
				Category: "Docker Plugin",
				Name:     "plugin-socket-user",
				EnvVars:  []string{constants.EnvVarsPrefix + "PLUGIN_SOCKET_USER"},
				Value:    currentUser.Username,
				Usage:    "Docker Plugin Unix socket user name or ID",
			},
			&cli.StringFlag{
				Category: "Docker Plugin",
				Name:     "plugin-socket-group",
				EnvVars:  []string{constants.EnvVarsPrefix + "PLUGIN_SOCKET_GROUP"},
				Value:    currentGroup.Name,
				Usage:    "Docker Plugin Unix socket group name or ID",
			},
			&cli.UintFlag{
				Category:    "Docker Plugin",
				Name:        "plugin-socket-mode",
				EnvVars:     []string{constants.EnvVarsPrefix + "PLUGIN_SOCKET_MODE"},
				Value:       0600,
				DefaultText: "0600",
				Usage:       "Docker Plugin Unix socket access modes",
			},
			&cli.BoolFlag{
				Category: "Docker Volume Driver",
				Name:     "disable-volume-driver",
				EnvVars:  []string{constants.EnvVarsPrefix + "DISABLE_VOLUME_DRIVER"},
				Value:    false,
				Usage:    "Disable Volume Driver",
			},
			&cli.BoolFlag{
				Category: "Docker Volume Driver",
				Name:     "volume-driver-global-scope",
				EnvVars:  []string{constants.EnvVarsPrefix + "VOLUME_DRIVER_GLOBAL_SCOPE"},
				Value:    false,
				Usage:    "Use global scope instead of local scope",
			},
			&cli.StringFlag{
				Category: "Docker Volume Driver",
				Name:     "volume-driver-state-file",
				EnvVars:  []string{constants.EnvVarsPrefix + "VOLUME_DRIVER_STATE_FILE"},
				Value:    path.Join("/var/local", constants.AppName, "state.json"),
				Usage:    "Volume Driver state file",
			},
			&cli.StringFlag{
				Category: "Docker Volume Driver",
				Name:     "volume-driver-mount-dir",
				EnvVars:  []string{constants.EnvVarsPrefix + "VOLUME_DRIVER_MOUNT_DIR"},
				Value:    path.Join(dockerSdkPlugin.DefaultDockerRootDirectory, constants.DockerPluginId),
				Usage:    "Volume Driver FS mount directory path",
			},
			&cli.StringFlag{
				Category: "Docker Volume Driver",
				Name:     "volume-driver-mount-user",
				EnvVars:  []string{constants.EnvVarsPrefix + "VOLUME_DRIVER_MOUNT_USER"},
				Value:    currentUser.Username,
				Usage:    "Volume Driver FS mount user name or ID",
			},
			&cli.StringFlag{
				Category: "Docker Volume Driver",
				Name:     "volume-driver-mount-group",
				EnvVars:  []string{constants.EnvVarsPrefix + "VOLUME_DRIVER_MOUNT_GROUP"},
				Value:    currentGroup.Name,
				Usage:    "Volume Driver FS mount group name or ID",
			},
			&cli.BoolFlag{
				Category: "Docker Secret Provider",
				Name:     "disable-secret-provider",
				EnvVars:  []string{constants.EnvVarsPrefix + "DISABLE_SECRET_PROVIDER"},
				Value:    false,
				Usage:    "Disable Secret Provider",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.IsSet("auth-mount") {
				v := ctx.String("auth-mount")
				defaultOptDocker.Secret.Vault.VaultAuth.MountPath = &v
			}

			if ctx.IsSet("engine-mount") {
				v := ctx.String("engine-mount")
				defaultOptDocker.Secret.Vault.VaultEngine.MountPath = &v
			}

			return start(ctx, defaultOptDocker)
		},
	}

	if err := app.Run(os.Args); err != nil {
		util.Fatalf("%v\n", err)
	}
}

func start(ctx *cli.Context, defaultOptDocker options.OptDocker) error {
	var arg string

	util.Printf("Plugin starting. Version: %s\n", ctx.App.Version)

	arg = ctx.String("plugin-socket-user")
	unixSocketUId, err := util.UserIdFromUser(arg)
	if err != nil {
		return fmt.Errorf("unable to find user %s: %w", arg, err)
	}

	arg = ctx.String("plugin-socket-group")
	unixSocketGId, err := util.UserGroupIdFromUserGroup(arg)
	if err != nil {
		return fmt.Errorf("unable to find group %s: %w", arg, err)
	}

	arg = ctx.String("volume-driver-mount-user")
	mountDirUId, err := util.UserIdFromUser(arg)
	if err != nil {
		return fmt.Errorf("unable to find user %s: %w", arg, err)
	}

	arg = ctx.String("volume-driver-mount-group")
	mountDirGId, err := util.UserGroupIdFromUserGroup(arg)
	if err != nil {
		return fmt.Errorf("unable to find group %s: %w", arg, err)
	}

	tcpBindAddr := ctx.String("plugin-tcp-bind-addr")

	var tcpBindPort *uint16
	ptbp := ctx.Uint("plugin-tcp-bind-port")
	if ptbp != 0 {
		tbp16 := uint16(ptbp)
		tcpBindPort = &tbp16
	}

	unixSocketPath := ctx.String("plugin-socket-path")

	dockerPlugin, err := docker.NewPlugin(docker.PluginConfig{
		TcpBindAddr:    &tcpBindAddr,
		TcpBindPort:    tcpBindPort,
		TcpTlsConfig:   nil, // TODO
		UnixSocketPath: &unixSocketPath,
		UnixSocketUId:  unixSocketUId,
		UnixSocketGId:  unixSocketGId,
		UnixSocketMode: uint32(ctx.Uint("plugin-socket-mode")),

		VolumeDriverDisabled:      ctx.Bool("disable-volume-driver"),
		VolumeDriverGlobalScope:   ctx.Bool("volume-driver-global-scope"),
		VolumeDriverStateFilePath: ctx.String("volume-driver-state-file"),
		VolumeDriverFsConfig: docker.FsConfig{
			MountFuseName: constants.AppName,
			MountDir:      ctx.String("volume-driver-mount-dir"),
			MountDirUId:   mountDirUId,
			MountDirGId:   mountDirGId,
		},

		SecretProviderDisabled: ctx.Bool("disable-secret-provider"),

		DefaultOptDocker: defaultOptDocker,
	})
	if err != nil {
		return fmt.Errorf("create plugin: %w", err)
	}

	if !ctx.Bool("disable-mlock") {
		if err := util.LockMemory(); err != nil {
			return fmt.Errorf("lock memory: %w", err)
		}
	}

	if err := dockerPlugin.Initialize(); err != nil {
		return fmt.Errorf("initialize plugin: %w", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	util.Printf("Started.\n")

	select {
	case <-sigs:
	case <-dockerPlugin.DoneChan():
	}

	util.Printf("Exiting...\n")

	dockerPlugin.CleanUp()
	return nil
}
