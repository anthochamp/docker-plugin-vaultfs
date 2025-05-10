// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package backendVault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/anthochamp/docker-plugin-vaultfs/options"
	"github.com/anthochamp/docker-plugin-vaultfs/util"
	"github.com/google/uuid"
	vaultApi "github.com/hashicorp/vault/api"
	vaultApiAuthApprole "github.com/hashicorp/vault/api/auth/approle"
	vaultApiAuthUserpass "github.com/hashicorp/vault/api/auth/userpass"
)

var (
	clientsCacheLock = &sync.Mutex{}
	clientsCache     = map[string]*VaultClient{}
)

type VaultClient struct {
	config VaultClientConfig

	refCounter int

	loginLock             *sync.Mutex
	client                *vaultApi.Client
	authLifetimeWatcherId *string

	lifetimeWatchersLock *sync.Mutex
	lifetimeWatchers     map[string]*vaultApi.LifetimeWatcher
}

type VaultClientConfig struct {
	optClientHttp options.OptClientHttp
	optVaultAuth  options.OptVaultAuth
}

func (z *VaultClientConfig) cacheId_() string {
	r := ""
	r += z.optClientHttp.CacheId_()
	r += z.optVaultAuth.CacheId_()
	return r
}

func newVaultClient(config VaultClientConfig) (*VaultClient, error) {
	util.Tracef("newVaultClient(%+v)\n", config)

	clientsCacheLock.Lock()
	defer clientsCacheLock.Unlock()

	cacheId := config.cacheId_()

	client, ok := clientsCache[cacheId]
	if ok {
		util.Tracef("Reusing client %+v\n", client)
		client.refCounter++
		return client, nil
	}

	if config.optClientHttp.Address == "" {
		return nil, errors.New("vault address must be defined")
	}

	client = &VaultClient{
		config: config,

		refCounter: 1,

		loginLock: &sync.Mutex{},

		lifetimeWatchersLock: &sync.Mutex{},
		lifetimeWatchers:     map[string]*vaultApi.LifetimeWatcher{},
	}

	util.Tracef("Creating new client %+v\n", client)
	clientsCache[cacheId] = client

	return client, nil
}

func (z *VaultClient) Close() {
	util.Tracef("VaultClient[%v].Close()\n", z)

	clientsCacheLock.Lock()
	defer clientsCacheLock.Unlock()

	z.refCounter--
	if z.refCounter == 0 {
		delete(clientsCache, z.config.cacheId_())

		z.logout()

		z.lifetimeWatchersLock.Lock()
		lifetimeWatchers := z.lifetimeWatchers
		z.lifetimeWatchers = map[string]*vaultApi.LifetimeWatcher{}
		z.lifetimeWatchersLock.Unlock()

		for _, v := range lifetimeWatchers {
			v.Stop()
		}
	}
}

func (z *VaultClient) createApiUnsafe(clientCertFile *string, clientKeyFile *string) error {
	apiConfig := vaultApi.DefaultConfig()
	apiConfig.Address = z.config.optClientHttp.Address
	apiConfig.DisableRedirects = z.config.optClientHttp.DisableRedirects

	var clientCert = ""
	var clientKey = ""
	var caCert = ""
	var tlsServerName = ""
	if clientCertFile != nil {
		clientCert = *clientCertFile
	} else if z.config.optClientHttp.Tls.CertFile != nil {
		clientCert = *z.config.optClientHttp.Tls.CertFile
	}
	if clientKeyFile != nil {
		clientKey = *clientKeyFile
	} else if z.config.optClientHttp.Tls.KeyFile != nil {
		clientKey = *z.config.optClientHttp.Tls.KeyFile
	}
	if z.config.optClientHttp.Tls.CACertFile != nil {
		caCert = *z.config.optClientHttp.Tls.CACertFile
	}
	if z.config.optClientHttp.Tls.ServerName != nil {
		tlsServerName = *z.config.optClientHttp.Tls.ServerName
	}

	if err := apiConfig.ConfigureTLS(&vaultApi.TLSConfig{
		Insecure:      z.config.optClientHttp.Tls.Insecure,
		ClientCert:    clientCert,
		ClientKey:     clientKey,
		CACert:        caCert,
		TLSServerName: tlsServerName,
	}); err != nil {
		return fmt.Errorf("configure TLS: %w", err)
	}

	apiClient, err := vaultApi.NewClient(apiConfig)
	if err != nil {
		return err
	}

	// ensure client didn't take infos from environment variables
	apiClient.ClearToken()
	apiClient.ClearNamespace()

	z.client = apiClient
	return nil
}

func (z *VaultClient) login() error {
	util.Tracef("VaultClient[%v].login()\n", z)

	z.loginLock.Lock()
	defer z.loginLock.Unlock()

	if z.client != nil {
		return nil
	}

	var authMethod vaultApi.AuthMethod
	var authSecret *vaultApi.Secret

	switch z.config.optVaultAuth.Method {
	case options.VaultAuthMethodAppRole:
		var roleId string
		if z.config.optVaultAuth.RoleIdFile == nil {
			roleId = *z.config.optVaultAuth.RoleId
		} else {
			content, err := os.ReadFile(*z.config.optVaultAuth.RoleIdFile)
			if err != nil {
				return fmt.Errorf("read RoleID file %s: %w", *z.config.optVaultAuth.RoleIdFile, err)
			}

			roleId = string(content)
		}

		var secretId *vaultApiAuthApprole.SecretID
		if z.config.optVaultAuth.SecretIdFile == nil {
			secretId = &vaultApiAuthApprole.SecretID{
				FromString: *z.config.optVaultAuth.SecretId,
			}
		} else {
			secretId = &vaultApiAuthApprole.SecretID{
				FromFile: *z.config.optVaultAuth.SecretIdFile,
			}
		}

		loginOptionMountPath := vaultApiAuthApprole.WithMountPath(z.config.optVaultAuth.EffectiveMountPath())

		var err error
		if z.config.optVaultAuth.SecretIdTokenWrapped {
			authMethod, err = vaultApiAuthApprole.NewAppRoleAuth(
				roleId, secretId, loginOptionMountPath,
				vaultApiAuthApprole.WithWrappingToken(),
			)
		} else {
			authMethod, err = vaultApiAuthApprole.NewAppRoleAuth(roleId, secretId, loginOptionMountPath)
		}

		if err != nil {
			return fmt.Errorf("create AppRole auth: %w", err)
		}

	case options.VaultAuthMethodCert:
		if err := z.createApiUnsafe(z.config.optVaultAuth.CertFile, z.config.optVaultAuth.CertKeyFile); err != nil {
			return fmt.Errorf("create api: %w", err)
		}

		path := fmt.Sprintf("auth/%s/login", z.config.optVaultAuth.EffectiveMountPath())

		var err error
		authSecret, err = z.client.Logical().Write(path, map[string]interface{}{})
		if err != nil {
			return fmt.Errorf("login with certificate: %w", err)
		}

	case options.VaultAuthMethodToken:
		var token string
		if z.config.optVaultAuth.TokenFile == nil {
			token = *z.config.optVaultAuth.Token
		} else {
			content, err := os.ReadFile(*z.config.optVaultAuth.TokenFile)
			if err != nil {
				return fmt.Errorf("read token file %s: %w", *z.config.optVaultAuth.TokenFile, err)
			}

			token = string(content)
		}

		if err := z.createApiUnsafe(nil, nil); err != nil {
			return fmt.Errorf("create api: %w", err)
		}

		z.client.SetToken(token)

	case options.VaultAuthMethodUserPass:
		var username string
		if z.config.optVaultAuth.UsernameFile == nil {
			username = *z.config.optVaultAuth.Username
		} else {
			content, err := os.ReadFile(*z.config.optVaultAuth.UsernameFile)
			if err != nil {
				return fmt.Errorf("read username file %s: %w", *z.config.optVaultAuth.UsernameFile, err)
			}

			username = string(content)
		}

		var password *vaultApiAuthUserpass.Password
		if z.config.optVaultAuth.SecretIdFile == nil {
			password = &vaultApiAuthUserpass.Password{
				FromString: *z.config.optVaultAuth.Password,
			}
		} else {
			password = &vaultApiAuthUserpass.Password{
				FromFile: *z.config.optVaultAuth.PasswordFile,
			}
		}

		userpassAuth, err := vaultApiAuthUserpass.NewUserpassAuth(
			username,
			password,
			vaultApiAuthUserpass.WithMountPath(z.config.optVaultAuth.EffectiveMountPath()),
		)
		if err != nil {
			return fmt.Errorf("create Userpass auth: %w", err)
		}

		authMethod = userpassAuth

	default:
		return errors.New("not implemented")
	}

	if authSecret == nil && authMethod != nil {
		if err := z.createApiUnsafe(nil, nil); err != nil {
			return fmt.Errorf("create api: %w", err)
		}

		// will set z.client.Token upon successful completion
		authSecret, err := z.client.Auth().Login(context.Background(), authMethod)
		if err != nil {
			return fmt.Errorf("login: %w", err)
		}

		if authSecret == nil {
			return fmt.Errorf("login did not return a token")
		}
	}

	if z.client == nil {
		return fmt.Errorf("internal error")
	}

	if authSecret != nil && authSecret.Auth.Renewable {
		lifetimeWatcherId, err := z.NewLifetimeWatcher(vaultApi.LifetimeWatcherInput{
			Secret:    authSecret,
			Increment: z.config.optVaultAuth.TokenRenewTtl,
		}, func(err error) {
			z.logout()

			if err != nil {
				util.Errorf("Unable to renew vault client %v auth secret: %v\n", z, err)
			}

		}, func(renewal *vaultApi.RenewOutput) {
			util.Tracef("Renewed vault client %v auth secret: %+v\n", z, renewal)
		})

		z.authLifetimeWatcherId = lifetimeWatcherId

		if err != nil {
			return fmt.Errorf("create lifetime watcher: %w", err)
		}
	}

	return nil
}

func (z *VaultClient) logout() {
	util.Tracef("VaultClient[%v].logout()\n", z)

	z.loginLock.Lock()
	defer z.loginLock.Unlock()

	if z.authLifetimeWatcherId != nil {
		z.CloseLifetimeWatcher(*z.authLifetimeWatcherId)
		z.authLifetimeWatcherId = nil
	}

	z.client = nil
}

func (z *VaultClient) NewLifetimeWatcher(input vaultApi.LifetimeWatcherInput, onDone func(error), onRenewed func(*vaultApi.RenewOutput)) (*string, error) {
	lifetimeWatcherId := uuid.New().String()

	lifetimeWatcher, err := z.client.NewLifetimeWatcher(&input)
	if err != nil {
		return nil, err
	}

	go lifetimeWatcher.Start()

	go func() {
		for {
			select {
			case err := <-lifetimeWatcher.DoneCh():
				if onDone != nil {
					onDone(err)
				}

				z.lifetimeWatchersLock.Lock()
				delete(z.lifetimeWatchers, lifetimeWatcherId)
				z.lifetimeWatchersLock.Unlock()
				return

			case renewal := <-lifetimeWatcher.RenewCh():
				if onRenewed != nil {
					onRenewed(renewal)
				}
			}
		}
	}()

	z.lifetimeWatchersLock.Lock()
	z.lifetimeWatchers[lifetimeWatcherId] = lifetimeWatcher
	z.lifetimeWatchersLock.Unlock()

	return &lifetimeWatcherId, nil
}

func (z *VaultClient) CloseLifetimeWatcher(id string) error {
	z.lifetimeWatchersLock.Lock()
	lifetimeWatcher, ok := z.lifetimeWatchers[id]
	z.lifetimeWatchersLock.Unlock()

	if !ok {
		return os.ErrNotExist
	}

	lifetimeWatcher.Stop()
	return nil
}

func (z *VaultClient) FetchKVv1Secret(engineMountPath string, secretPath string) (*vaultApi.KVSecret, error) {
	util.Tracef("VaultClient[%v].getKVv1Secret(%s, %s)\n", z, engineMountPath, secretPath)

	if err := z.login(); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	vaultKvSecret, err := z.client.KVv1(engineMountPath).Get(context.Background(), secretPath)
	if err != nil {
		if err == vaultApi.ErrSecretNotFound {
			return nil, os.ErrNotExist
		}

		// until we got a better way of detecting it, just logout on error (auth token might be expired)
		z.logout()

		return nil, err
	}

	return vaultKvSecret, nil
}

func (z *VaultClient) FetchKVv2Secret(engineMountPath string, secretPath string, secretVersion *int) (*vaultApi.KVSecret, error) {
	util.Tracef("VaultClient[%v].getKVv2Secret(%s, %s, %v)\n", z, engineMountPath, secretPath, secretVersion)

	if err := z.login(); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	var vaultKvSecret *vaultApi.KVSecret
	var err error

	if secretVersion == nil {
		vaultKvSecret, err = z.client.KVv2(engineMountPath).Get(context.Background(), secretPath)
	} else {
		vaultKvSecret, err = z.client.KVv2(engineMountPath).GetVersion(context.Background(), secretPath, *secretVersion)
	}

	if err != nil {
		if err == vaultApi.ErrSecretNotFound {
			return nil, os.ErrNotExist
		}

		// until we got a better way of detecting it, just logout on error (auth token might be expired)
		z.logout()

		return nil, err
	}

	return vaultKvSecret, nil
}
