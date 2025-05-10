// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"net/http"

	"github.com/anthochamp/docker-plugin-vaultfs/util"
)

type VolumeDriver interface {
	Create(VolumeDriverCreateRequest) error
	List() (*VolumeDriverListResponse, error)
	Get(VolumeDriverGetRequest) (*VolumeDriverGetResponse, error)
	Remove(VolumeDriverRemoveRequest) error
	Path(VolumeDriverPathRequest) (*VolumeDriverPathResponse, error)
	Mount(VolumeDriverMountRequest) (*VolumeDriverMountResponse, error)
	Unmount(VolumeDriverUnmountRequest) error
	Capabilities() (*VolumeDriverCapabilitiesResponse, error)
}

func RegisterVolumeDriver(vd VolumeDriver, p Plugin) {
	p.RegisterHandlerFunc(volumeDriverCreatePath, func(hr util.HttpRequest) error { return handleVolumeDriverCreate(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverListPath, func(hr util.HttpRequest) error { return handleVolumeDriverList(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverGetPath, func(hr util.HttpRequest) error { return handleVolumeDriverGet(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverRemovePath, func(hr util.HttpRequest) error { return handleVolumeDriverRemove(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverPathPath, func(hr util.HttpRequest) error { return handleVolumeDriverPath(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverMountPath, func(hr util.HttpRequest) error { return handleVolumeDriverMount(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverUnmountPath, func(hr util.HttpRequest) error { return handleVolumeDriverUnmount(vd, hr) })
	p.RegisterHandlerFunc(volumeDriverCapabilitiesPath, func(hr util.HttpRequest) error { return handleVolumeDriverCapabilities(vd, hr) })

	p.Manifest.Implements = append(p.Manifest.Implements, pluginManifestVolumeDriverImplementId)
}

func handleVolumeDriverCreate(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverCreateRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	if err := vd.Create(req); err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(struct{}{})
}

func handleVolumeDriverList(vd VolumeDriver, hr util.HttpRequest) error {
	res, err := vd.List()
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(res)
}

func handleVolumeDriverGet(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverGetRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	res, err := vd.Get(req)
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(res)
}

func handleVolumeDriverRemove(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverRemoveRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	err := vd.Remove(req)
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(struct{}{})
}

func handleVolumeDriverPath(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverPathRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	res, err := vd.Path(req)
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(res)
}

func handleVolumeDriverMount(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverMountRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	res, err := vd.Mount(req)
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(res)
}

func handleVolumeDriverUnmount(vd VolumeDriver, hr util.HttpRequest) error {
	var req VolumeDriverUnmountRequest
	if err := hr.DecodeJsonBody(&req); err != nil {
		return hr.HttpErrorStr(http.StatusBadRequest, err.Error())
	}

	if err := vd.Unmount(req); err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(struct{}{})
}

func handleVolumeDriverCapabilities(vd VolumeDriver, hr util.HttpRequest) error {
	res, err := vd.Capabilities()
	if err != nil {
		return hr.HttpErrorJson(http.StatusInternalServerError, ErrorResponse{Err: err.Error()})
	}

	return hr.WriteJson(res)
}
