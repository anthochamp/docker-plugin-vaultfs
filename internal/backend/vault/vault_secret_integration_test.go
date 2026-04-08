// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

//go:build integration

package backendVault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/anthochamp/docker-plugin-vaultfs/internal/constants"
	"github.com/anthochamp/docker-plugin-vaultfs/internal/options"
	vaultApi "github.com/hashicorp/vault/api"
)

const (
	integrationKVv1Mount   = "integration-kv1"
	integrationKVv2Mount   = "integration-kv2"
	integrationSecretPath  = "test-secret"
	integrationCachedPath  = "test-cached-secret"
	integrationMissingPath = "does-not-exist"
)

var integrationVaultAddr string
var integrationVaultToken string

func TestMain(m *testing.M) {
	integrationVaultAddr = os.Getenv("VAULT_ADDR")
	if integrationVaultAddr == "" {
		integrationVaultAddr = "http://127.0.0.1:8200"
	}

	integrationVaultToken = os.Getenv("VAULT_TOKEN")
	if integrationVaultToken == "" {
		integrationVaultToken = "myroot"
	}

	if err := integrationSetup(); err != nil {
		fmt.Fprintf(os.Stderr, "integration test setup failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func integrationSetup() error {
	adminClient, err := newAdminClient()
	if err != nil {
		return fmt.Errorf("create admin client: %w", err)
	}

	if err = adminClient.Sys().Mount(integrationKVv1Mount, &vaultApi.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "1"},
	}); err != nil {
		return fmt.Errorf("mount kv-v1 engine: %w", err)
	}

	if err = adminClient.Sys().Mount(integrationKVv2Mount, &vaultApi.MountInput{
		Type:    "kv",
		Options: map[string]string{"version": "2"},
	}); err != nil {
		return fmt.Errorf("mount kv-v2 engine: %w", err)
	}

	secretData := map[string]interface{}{
		"username": "admin",
		"password": "hunter2",
	}

	if err = adminClient.KVv1(integrationKVv1Mount).Put(context.Background(), integrationSecretPath, secretData); err != nil {
		return fmt.Errorf("write KV v1 secret: %w", err)
	}

	if _, err = adminClient.KVv2(integrationKVv2Mount).Put(context.Background(), integrationSecretPath, secretData); err != nil {
		return fmt.Errorf("write KV v2 secret: %w", err)
	}

	if _, err = adminClient.KVv2(integrationKVv2Mount).Put(context.Background(), integrationCachedPath, secretData); err != nil {
		return fmt.Errorf("write KV v2 cached secret: %w", err)
	}

	if err = adminClient.KVv2(integrationKVv2Mount).PutMetadata(context.Background(), integrationCachedPath, vaultApi.KVMetadataPutInput{
		CustomMetadata: map[string]interface{}{
			constants.AppName + "-cache-ttl": "60",
		},
	}); err != nil {
		return fmt.Errorf("write KV v2 custom metadata: %w", err)
	}

	return nil
}

func newAdminClient() (*vaultApi.Client, error) {
	config := vaultApi.DefaultConfig()
	config.Address = integrationVaultAddr

	client, err := vaultApi.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(integrationVaultToken)

	return client, nil
}

func newVaultSecretForTest(kvVersion int, mountPath string, secretPath string) (*VaultSecret, error) {
	return NewVaultSecret(VaultSecretConfig{
		OptVault: options.OptVault{
			ClientHttp: options.OptClientHttp{
				Address: integrationVaultAddr,
			},
			VaultAuth: options.OptVaultAuth{
				Method: options.VaultAuthMethodToken,
				Token:  &integrationVaultToken,
			},
			VaultEngine: options.OptVaultEngine{
				Type:      options.VaultEngineTypeKv,
				KvVersion: kvVersion,
				MountPath: &mountPath,
			},
			VaultSecret: options.OptVaultSecret{
				Path: secretPath,
			},
		},
	})
}

func TestVaultSecretGetDataKVv1(t *testing.T) {
	secret, err := newVaultSecretForTest(1, integrationKVv1Mount, integrationSecretPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	data, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil data")
	}

	usernameValue, ok := (*data).GetValue("username")
	if !ok || usernameValue == nil || *usernameValue != "admin" {
		t.Errorf("expected username=admin, got %v", usernameValue)
	}

	passwordValue, ok := (*data).GetValue("password")
	if !ok || passwordValue == nil || *passwordValue != "hunter2" {
		t.Errorf("expected password=hunter2, got %v", passwordValue)
	}
}

func TestVaultSecretGetDataKVv2(t *testing.T) {
	secret, err := newVaultSecretForTest(2, integrationKVv2Mount, integrationSecretPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	data, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil data")
	}

	usernameValue, ok := (*data).GetValue("username")
	if !ok || usernameValue == nil || *usernameValue != "admin" {
		t.Errorf("expected username=admin, got %v", usernameValue)
	}

	// KV v2 automatically adds .version-metadata-* keys
	keys := (*data).GetKeys()
	hasVersionMetadata := false
	for _, key := range keys {
		if key == ".version-metadata-version" {
			hasVersionMetadata = true
			break
		}
	}

	if !hasVersionMetadata {
		t.Error("expected .version-metadata-version key from KV v2 version metadata")
	}
}

func TestVaultSecretGetDataMissingSecret(t *testing.T) {
	secret, err := newVaultSecretForTest(1, integrationKVv1Mount, integrationMissingPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	_, err = secret.GetData(false)

	if err == nil {
		t.Fatal("expected error for missing secret")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestVaultSecretGetDataCachesResult(t *testing.T) {
	secret, err := newVaultSecretForTest(1, integrationKVv1Mount, integrationSecretPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	firstData, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("first GetData failed: %v", err)
	}

	secondData, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("second GetData failed: %v", err)
	}

	if (*firstData).UniqueId() != (*secondData).UniqueId() {
		t.Error("expected second call to return cached data with the same UniqueId")
	}
}

func TestVaultSecretGetDataNoCacheBypassesCache(t *testing.T) {
	secret, err := newVaultSecretForTest(1, integrationKVv1Mount, integrationSecretPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	firstData, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("first GetData failed: %v", err)
	}

	secondData, err := secret.GetData(true)
	if err != nil {
		t.Fatalf("second GetData (noCache=true) failed: %v", err)
	}

	if (*firstData).UniqueId() == (*secondData).UniqueId() {
		t.Error("expected noCache=true to fetch fresh data with a different UniqueId")
	}
}

func TestVaultSecretGetDataCustomCacheTTL(t *testing.T) {
	secret, err := newVaultSecretForTest(2, integrationKVv2Mount, integrationCachedPath)
	if err != nil {
		t.Fatalf("failed to create VaultSecret: %v", err)
	}
	defer secret.Close()

	data, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil data")
	}

	// verify the custom cache-ttl metadata key is consumed internally and not exposed as a data field
	_, hasCacheTtlKey := (*data).GetValue(constants.AppName + "-cache-ttl")
	if hasCacheTtlKey {
		t.Errorf("expected %s-cache-ttl to be consumed as cache config, not exposed as a data field", constants.AppName)
	}

	// verify the cache is active: second call should return the same UniqueId
	secondData, err := secret.GetData(false)
	if err != nil {
		t.Fatalf("second GetData failed: %v", err)
	}

	if (*data).UniqueId() != (*secondData).UniqueId() {
		t.Error("expected cached result with the same UniqueId for a secret with cache-ttl set")
	}
}
