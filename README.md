# Docker plugin for Hashicorp Vault

<!-- TOC tocDepth:2..4 chapterDepth:2..6 -->

- [Installation](#installation)
  - [As a Docker plugin (recommended)](#as-a-docker-plugin-recommended)
  - [As an external program](#as-an-external-program)
- [Usage](#usage)
  - [Docker Volume driver](#docker-volume-driver)
  - [More examples](#more-examples)
    - [K/V v1 example](#kv-v1-example)
    - [K/V v2 example](#kv-v2-example)
- [References](#references)
  - [Authentication Methods](#authentication-methods)
    - [AppRole](#approle)
    - [TLS certificates](#tls-certificates)
    - [Token](#token)
    - [Username/password](#usernamepassword)
  - [Secrets engines](#secrets-engines)
    - [Common options](#common-options)
    - [Key/Value engine](#keyvalue-engine)
    - [Database engines](#database-engines)
    - [PKI engine](#pki-engine)
- [Development](#development)
  - [Compilation](#compilation)
    - [As a local binary file](#as-a-local-binary-file)
    - [As a docker image](#as-a-docker-image)
    - [As a docker plugin](#as-a-docker-plugin)
  - [Development rationals](#development-rationals)
- [Notes](#notes)
  - [Rationals on Docker Plugin capabilities requirement](#rationals-on-docker-plugin-capabilities-requirement)
  - [Docker limitation on Volume options inheritances](#docker-limitation-on-volume-options-inheritances)

<!-- /TOC -->

## Installation

### As a Docker plugin (recommended)

```shell
docker plugin install anthochamp/vaultfs-amd64 --alias vaultfs

# check current environment 
docker plugin inspect vaultfs

# change environment
docker plugin set vaultfs DPV_VAULT_URL="https://your-vault-host:8200" 

docker plugin enable vaultfs

# start creating volumes...
```

### As an external program

- Compile the [binary version](#as-a-local-binary-file) of the program
- On a systemd compatible system, use `packages/systemd/docker-plugin-vault.service`
and `packages/systemd/docker-plugin-vault.socket`.

## Usage

### Docker Volume driver

You want a volume `secrets`, bound to the 4th version of your `credentials` secret
stored in a KV v2 engine mounted on `app`:

```shell
docker volume create \
    --driver vaultfs \
    -o auth-method=token
    -o token=<token>
    -o engine-mount=app
    -o engine-type=kv
    -o kv-engine-version=2
    -o secret=credentials
    -o kv-secret-version=4
    secrets

docker run -it --volume secrets:/run/secrets alpine sh
```

Given that the 4th version of your `credentials` secrets has two keys `username`
and `password`, you'll find two files `/run/mysecret/username` and `/run/mysecret/password`
containg the raw values of the associated Vault Secret keys.

Provided you didn't customized the default values from the command-line or the Docker
plugin environment, the above command can be simplified as:

```shell
docker volume create \
    --driver vaultfs \
    -o token=<token>
    -o engine-mount=app
    credentials@4

docker run -it --volume credentials@4:/run/secrets alpine sh
```

### More examples

Minimal example for generating a lease for credentials for the role `public` in
the `database` engine:

```shell
docker volume create \
    --driver vaultfs \
    -o secret=public
    mycredentials
```

The plugin will automatically renew the lease or request a new lease if renewing
is refused.

#### K/V v1 example

Example for the secret `credentials` in the `app`:

```shell
docker volume create \
    --driver vaultfs \
    -o engine-type=kv
    -o engine=kv
    -o kv-engine-version=1
    -o secret=credentials
    mycredentials
```

Alternative example if the Docker Volume name is the same as the Vault Secret path:

```shell
docker volume create \
    --driver vaultfs \
    -o engine-type=kv
    -o engine=app
    -o kv-engine-version=1
    credentials
```

#### K/V v2 example

Example for the **4th version** of the secret **credentials** in the secret **app**:

```shell
docker volume create \
    --driver vaultfs \
    -o engine-type=kv
    -o engine=app
    -o kv-engine-version=2
    -o secret=credentials
    -o kv-secret-version=4
    mycredentials
```

Alternative example if the Docker Volume name is the same as the Vault Secret path:

```shell
docker volume create \
    --driver vaultfs \
    -o engine-type=kv
    -o engine=app
    -o kv-engine-version=2
    credentials@4
```

## References

> **Notes**: The default values of each fields can be changed using Docker plugin
> options.

### Authentication Methods

The following [Vault authentication methods](https://developer.hashicorp.com/vault/docs/auth)
are supported:

| `auth-method` | Documentation | Default `auth-mount` value
| - | - | -
| `approle` | [AppRole](#approle) | `approle`
| `cert` | [SSL/TLS certificates](#tls-certificates) | `cert`
| `token` | [Token](#token) | `token`
| `userpass` | [Username / Password](#usernamepassword) | `userpass`

All the authentication methods supports the following Docker Volume options:

| Volume option | Default value | Description
| - | - | -
| `auth-method` | `token` | Vault auth method id (case-insensitive)
| `auth-mount` | Based on `auth-method`, see table above | Path to the Vault auth method
| `auth-token-renew-ttl` | `0` | The authentication token TTL (in seconds) to request to the engine.

#### AppRole

[Vault AppRole](https://developer.hashicorp.com/vault/docs/auth/approle) authentication
method, selectable via `auth-method=approle`, supports the following additional Docker
Volume options:

| Volume option | Default value | Description
| - | - | -
| `auth-role-id` | *none* | The RoleID to use for authentication
| `auth-role-id-file` | *none* | The path to a file containing the RoleID to use for authentication
| `auth-secret-id` | *none* | The SecretID to use for authentication
| `auth-secret-id-file` | *none* | The path to a file containing the SecretID to use for authentication

cf. <https://developer.hashicorp.com/vault/docs/auth/approle#code-example>

Example:

```shell
docker volume create \
    --driver vaultfs \
    -o secret=credentials
    -o auth-method=approle
    -o auth-role-id=<role-id>
    -o auth-secret-id=<secret-id>
    mycredentials
```

#### TLS certificates

[Vault TLS certificates](https://developer.hashicorp.com/vault/docs/auth/cert)
authentication method, selectable via `auth-method=cert`, supports the following
additional Docker Volume options:

| Volume option | Default value | Description
| - | - | -
| `auth-cert-file` | *none* | The path to the client certificate to use for authentication
| `auth-cert-key-file` | *none* | The path to the client certificate's key to use for authentication

Example:

```shell
docker volume create \
    --driver vaultfs \
    -o secret=credentials
    -o auth-method=cert
    -o cert-file=/path/to/volume/cert.pem
    -o cert-key-file=/path/to/volume/key.pem
    mycredentials
```

#### Token

[Vault Token](https://developer.hashicorp.com/vault/docs/auth/token) authentication
method, selectable via `auth-method=token`, supports the following additional Docker
Volume options:

| Volume option | Default value | Description
| - | - | -
| `auth-token` | *none* | The token to use for authentication
| `auth-token-file` | *none* | The path to a file containing the token to use for authentication

Example:

```shell
docker volume create \
    --driver vaultfs \
    -o secret=credentials
    -o auth-method=token
    -o auth-token=<token>
    mycredentials
```

#### Username/password

[Vault Userpass](https://developer.hashicorp.com/vault/docs/auth/userpass) authentication
method, selectable via `auth-method=userpass`, supports the following additional
Docker Volume options:

| Volume option | Default value | Description
| - | - | -
| `username` | *none* | The username to use for authentication
| `username-file` | *none* | The path to a file containing the username to use for authentication
| `password` | *none* | The password to use for authentication
| `password-file` | *none* | The path to a file containing the password to use for authentication

Example:

```shell
docker volume create \
    --driver vaultfs \
    -o secret=credentials
    -o auth-method=userpass
    -o username=<username>
    -o password=<password>
    mycredentials
```

### Secrets engines

The following Vault Secrets engines[^1] are supported:

| Engine | `engine-type` | Default `engine-mount` value
| - | - | -
| [Key/Value](#keyvalue-engine) | `kv` | `secret`
| [Database](#database-engines) | `db` | `database`
| [PKI](#pki-engine) | `pki` | `pki`

#### Common options

All the above secrets engines support the following **Docker Volume** options:

| Volume option | Default value | Description
| - | - | -
| `engine-type` | `kv` | Vault engine internal type identifier (case-insensitive)
| `engine-mount` | Based on `engine-type`, see table above | Mount path to the Vault engine (Vault's CLI `-mount` equivalent)
| `secret` | *none* | Path to the secret inside the Vault engine
| `token-renew-ttl` | `0` | The secret token TTL (in seconds) to request to the engine
| `mount-uid` | `0` | User ID of the secret directory and its fields files
| `mount-gid` | `0` | Group ID of the secret directory and its fields files
| `mount-mode` | `0550` | Access mode of the secret directory
| `field-mount-mode` | `0440` | Access mode of the secret's fields files

#### Key/Value engine

Vault Key/Value engine[^2] supports the following additional **Docker Volume** options:

| Volume option | Default value | Description
| - | - | -
| `kv-engine-version` | `1` | K/V engine version (`1` or `2`)
| `kv-secret-version` | *none* | The version of the secret as an integer or `latest`. No value defined is the same as `latest`. Relevant only if `kv-engine-version=2`.

#### Database engines

Vault Database engines[^3] support the following additional **Docker Volume** options:

| Volume option | Default value | Description
| - | - | -
| `db-role` | *none* | Path to the database role

#### PKI engine

Vault PKI engine[^4] supports the following additional **Docker Volume** options:

| Volume option | Default value | Description
| - | - | -

[^1]: [Vault Secrets documentation (official)](https://developer.hashicorp.com/vault/docs/secrets)
[^2]: [Vault Key/Value engine documentation (official)](https://developer.hashicorp.com/vault/docs/secrets/kv)
[^3]: [Vault Databases engines documentation (official)](https://developer.hashicorp.com/vault/docs/secrets/databases)
[^4]: [Vault PKI engine documentation (official)](https://developer.hashicorp.com/vault/docs/secrets/pki)

## Development

### Compilation

Clone the repository.

#### As a local binary file

Install golang using the official documentation: <https://go.dev/doc/install>.

Build the plugin :

```shell
./scripts/build.sh

./src/docker-plugin-vaultfs --help
```

#### As a docker image

```shell
./scripts/docker-image/build.sh linux/amd64 vaultfs:latest

docker run -it --cap-add=IPC_LOCK --cap-add=SYS_ADMIN --device /dev/fuse --security-opt apparmor:unconfined vaultfs:latest
```

#### As a docker plugin

```shell
# requires sudo access
./scripts/docker-plugin/create.sh linux/amd64 vaultfs

docker plugin enable vaultfs
```

### Development rationals

- Q: Why not use docker/go-plugins-helpers ?
A: <https://github.com/docker/go-plugins-helpers/issues/123>

## Notes

### Rationals on Docker Plugin capabilities requirement

- `CAP_SYS_ADMIN` is required to mount/unmount the FUSE volume FS used for the
Docker Volume driver;
- `CAP_IPC_LOCK` is required to lock the memory so that Vault secrets stays in
memory and do not get leaked on swap.

Reference: <https://man7.org/linux/man-pages/man7/capabilities.7.html>

### Docker limitation on Volume options inheritances

```shell
docker plugin set vaultfs auth-method=token token=<tokenA>

docker volume create \
    --driver vaultfs \
    -o token=<tokenB>
    mysecret

docker run -it \
    --mount volume-driver=vaultfs,source=mysecret,target=/run/mysecret,volume-opt=token=<tokenC> \
    alpine sh
```

In the example above, all the volume options (`volume-opt`) used in the `--mount`
argument to run the alpine container won't be registered by the plugin.

This is a limitation of the Docker plugin protocol. Docker only sends volume options
if the volume hasn't been created yet.

In the above case, since the volume `mysecret` has already been created via the
`docker volume create` command, any following reference to that volume will get
its volume options ignored. Effectively, the alpine container will use `tokenB`
to access the Vault Secret.
