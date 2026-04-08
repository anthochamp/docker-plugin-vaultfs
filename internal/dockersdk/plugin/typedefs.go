// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

const (
	// DefaultDockerRootDirectory is the default directory where volumes will be created.
	DefaultDockerRootDirectory = "/var/lib/docker-volumes"

	applicationDockerPluginsJsonMimeType = "application/vnd.docker.plugins.v1.1+json"

	pluginActivatePath           = "/Plugin.Activate"
	secretProviderGetSecretPath  = "/SecretProvider.GetSecret"
	volumeDriverCreatePath       = "/VolumeDriver.Create"
	volumeDriverGetPath          = "/VolumeDriver.Get"
	volumeDriverListPath         = "/VolumeDriver.List"
	volumeDriverRemovePath       = "/VolumeDriver.Remove"
	volumeDriverPathPath         = "/VolumeDriver.Path"
	volumeDriverMountPath        = "/VolumeDriver.Mount"
	volumeDriverUnmountPath      = "/VolumeDriver.Unmount"
	volumeDriverCapabilitiesPath = "/VolumeDriver.Capabilities"

	pluginManifestVolumeDriverImplementId   = `VolumeDriver`
	pluginManifestSecretProviderImplementId = `secretprovider`
)

type PluginManifest struct {
	Implements []string `json:",omitempty"`
}

// Request is the plugin secret request
type SecretProviderGetSecretRequest struct {
	SecretName          string            `json:",omitempty"` // SecretName is the name of the secret to request from the plugin
	SecretLabels        map[string]string `json:",omitempty"` // SecretLabels capture environment names and other metadata pertaining to the secret
	ServiceHostname     string            `json:",omitempty"` // ServiceHostname is the hostname of the service, can be used for x509 certificate
	ServiceName         string            `json:",omitempty"` // ServiceName is the name of the service that requested the secret
	ServiceID           string            `json:",omitempty"` // ServiceID is the name of the service that requested the secret
	ServiceLabels       map[string]string `json:",omitempty"` // ServiceLabels capture environment names and other metadata pertaining to the service
	TaskID              string            `json:",omitempty"` // TaskID is the ID of the task that the secret is assigned to
	TaskName            string            `json:",omitempty"` // TaskName is the name of the task that the secret is assigned to
	TaskImage           string            `json:",omitempty"` // TaskName is the image of the task that the secret is assigned to
	ServiceEndpointSpec *EndpointSpec     `json:",omitempty"` // ServiceEndpointSpec holds the specification for endpoints
}

type SecretProviderGetSecretResponse struct {
	Value []byte // Value is the value of the secret

	// DoNotReuse indicates that the secret returned from this request should
	// only be used for one task, and any further tasks should call the secret
	// driver again.
	DoNotReuse bool
}

// Response contains the plugin secret value
type secretProviderGetSecretHttpResponse struct {
	Value []byte `json:",omitempty"` // Value is the value of the secret
	Err   string `json:",omitempty"` // Err is the error response of the plugin

	// DoNotReuse indicates that the secret returned from this request should
	// only be used for one task, and any further tasks should call the secret
	// driver again.
	DoNotReuse bool `json:",omitempty"`
}

// CreateRequest is the structure that docker's requests are deserialized to.
type VolumeDriverCreateRequest struct {
	Name    string
	Options map[string]string `json:"Opts,omitempty"`
}

// RemoveRequest structure for a volume remove request
type VolumeDriverRemoveRequest struct {
	Name string
}

// MountRequest structure for a volume mount request
type VolumeDriverMountRequest struct {
	Name string
	ID   string
}

// MountResponse structure for a volume mount response
type VolumeDriverMountResponse struct {
	Mountpoint string
}

// UnmountRequest structure for a volume unmount request
type VolumeDriverUnmountRequest struct {
	Name string
	ID   string
}

// PathRequest structure for a volume path request
type VolumeDriverPathRequest struct {
	Name string
}

// PathResponse structure for a volume path response
type VolumeDriverPathResponse struct {
	Mountpoint string
}

// GetRequest structure for a volume get request
type VolumeDriverGetRequest struct {
	Name string
}

// GetResponse structure for a volume get response
type VolumeDriverGetResponse struct {
	Volume *Volume
}

// ListResponse structure for a volume list response
type VolumeDriverListResponse struct {
	Volumes []*Volume
}

// CapabilitiesResponse structure for a volume capability response
type VolumeDriverCapabilitiesResponse struct {
	Capabilities VolumeDriverCapability
}

// EndpointSpec represents the spec of an endpoint.
type EndpointSpec struct {
	Mode  int32        `json:",omitempty"`
	Ports []PortConfig `json:",omitempty"`
}

// PortConfig represents the config of a port.
type PortConfig struct {
	Name     string `json:",omitempty"`
	Protocol int32  `json:",omitempty"`
	// TargetPort is the port inside the container
	TargetPort uint32 `json:",omitempty"`
	// PublishedPort is the port on the swarm hosts
	PublishedPort uint32 `json:",omitempty"`
	// PublishMode is the mode in which port is published
	PublishMode int32 `json:",omitempty"`
}

// Volume represents a volume object for use with `Get` and `List` requests
type Volume struct {
	Name       string
	Mountpoint string                 `json:",omitempty"`
	CreatedAt  string                 `json:",omitempty"`
	Status     map[string]interface{} `json:",omitempty"`
}

// Capability represents the list of capabilities a volume driver can return
type VolumeDriverCapability struct {
	Scope string
}

// ErrorResponse is a formatted error message that docker can understand
type ErrorResponse struct {
	Err string
}

// NewErrorResponse creates an ErrorResponse with the provided message
func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{Err: msg}
}
