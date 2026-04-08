// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dockerSdkPlugin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockVolumeDriver implements VolumeDriver with configurable per-call behaviour.
type mockVolumeDriver struct {
	createFn       func(VolumeDriverCreateRequest) error
	listFn         func() (*VolumeDriverListResponse, error)
	getFn          func(VolumeDriverGetRequest) (*VolumeDriverGetResponse, error)
	removeFn       func(VolumeDriverRemoveRequest) error
	pathFn         func(VolumeDriverPathRequest) (*VolumeDriverPathResponse, error)
	mountFn        func(VolumeDriverMountRequest) (*VolumeDriverMountResponse, error)
	unmountFn      func(VolumeDriverUnmountRequest) error
	capabilitiesFn func() (*VolumeDriverCapabilitiesResponse, error)
}

func (m *mockVolumeDriver) Create(req VolumeDriverCreateRequest) error {
	if m.createFn != nil {
		return m.createFn(req)
	}
	return nil
}

func (m *mockVolumeDriver) List() (*VolumeDriverListResponse, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	return &VolumeDriverListResponse{}, nil
}

func (m *mockVolumeDriver) Get(req VolumeDriverGetRequest) (*VolumeDriverGetResponse, error) {
	if m.getFn != nil {
		return m.getFn(req)
	}
	return &VolumeDriverGetResponse{}, nil
}

func (m *mockVolumeDriver) Remove(req VolumeDriverRemoveRequest) error {
	if m.removeFn != nil {
		return m.removeFn(req)
	}
	return nil
}

func (m *mockVolumeDriver) Path(req VolumeDriverPathRequest) (*VolumeDriverPathResponse, error) {
	if m.pathFn != nil {
		return m.pathFn(req)
	}
	return &VolumeDriverPathResponse{}, nil
}

func (m *mockVolumeDriver) Mount(req VolumeDriverMountRequest) (*VolumeDriverMountResponse, error) {
	if m.mountFn != nil {
		return m.mountFn(req)
	}
	return &VolumeDriverMountResponse{}, nil
}

func (m *mockVolumeDriver) Unmount(req VolumeDriverUnmountRequest) error {
	if m.unmountFn != nil {
		return m.unmountFn(req)
	}
	return nil
}

func (m *mockVolumeDriver) Capabilities() (*VolumeDriverCapabilitiesResponse, error) {
	if m.capabilitiesFn != nil {
		return m.capabilitiesFn()
	}
	return &VolumeDriverCapabilitiesResponse{}, nil
}

// doVolumeDriverRequest sends a POST request to the given path on a plugin
// that has the provided VolumeDriver registered.
func doVolumeDriverRequest(driver VolumeDriver, path string, body string) *httptest.ResponseRecorder {
	plugin := MakePlugin()
	RegisterVolumeDriver(driver, plugin)

	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	recorder := httptest.NewRecorder()

	plugin.serveMux.ServeHTTP(recorder, request)

	return recorder
}

func TestVolumeDriverCreate(t *testing.T) {
	t.Run("happy path returns 200 with empty JSON object", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverCreatePath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		body := strings.TrimSpace(recorder.Body.String())
		if body != "{}" {
			t.Errorf("expected empty JSON object, got %q", body)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			createFn: func(_ VolumeDriverCreateRequest) error {
				return errors.New("create failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverCreatePath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(recorder.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp.Err != "create failed" {
			t.Errorf("expected Err=%q, got %q", "create failed", errResp.Err)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverCreatePath, `{invalid json}`)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverList(t *testing.T) {
	t.Run("happy path returns volume list", func(t *testing.T) {
		driver := &mockVolumeDriver{
			listFn: func() (*VolumeDriverListResponse, error) {
				return &VolumeDriverListResponse{
					Volumes: []*Volume{{Name: "vol1", Mountpoint: "/mnt/vol1"}},
				}, nil
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverListPath, `{}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response VolumeDriverListResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.Volumes) != 1 || response.Volumes[0].Name != "vol1" {
			t.Errorf("unexpected response: %+v", response)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			listFn: func() (*VolumeDriverListResponse, error) {
				return nil, errors.New("list failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverListPath, `{}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverGet(t *testing.T) {
	t.Run("happy path returns volume", func(t *testing.T) {
		driver := &mockVolumeDriver{
			getFn: func(req VolumeDriverGetRequest) (*VolumeDriverGetResponse, error) {
				return &VolumeDriverGetResponse{Volume: &Volume{Name: req.Name, Mountpoint: "/mnt/" + req.Name}}, nil
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverGetPath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response VolumeDriverGetResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Volume == nil || response.Volume.Name != "myvol" {
			t.Errorf("unexpected response: %+v", response)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			getFn: func(_ VolumeDriverGetRequest) (*VolumeDriverGetResponse, error) {
				return nil, errors.New("get failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverGetPath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverGetPath, `{invalid}`)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverRemove(t *testing.T) {
	t.Run("happy path returns 200 with empty JSON object", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverRemovePath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			removeFn: func(_ VolumeDriverRemoveRequest) error {
				return errors.New("remove failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverRemovePath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverPath(t *testing.T) {
	t.Run("happy path returns mountpoint", func(t *testing.T) {
		driver := &mockVolumeDriver{
			pathFn: func(_ VolumeDriverPathRequest) (*VolumeDriverPathResponse, error) {
				return &VolumeDriverPathResponse{Mountpoint: "/var/lib/docker-volumes/myvol"}, nil
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverPathPath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response VolumeDriverPathResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Mountpoint != "/var/lib/docker-volumes/myvol" {
			t.Errorf("unexpected mountpoint: %q", response.Mountpoint)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			pathFn: func(_ VolumeDriverPathRequest) (*VolumeDriverPathResponse, error) {
				return nil, errors.New("path failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverPathPath, `{"Name":"myvol"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverMount(t *testing.T) {
	t.Run("happy path returns mountpoint", func(t *testing.T) {
		driver := &mockVolumeDriver{
			mountFn: func(_ VolumeDriverMountRequest) (*VolumeDriverMountResponse, error) {
				return &VolumeDriverMountResponse{Mountpoint: "/var/lib/docker-volumes/myvol"}, nil
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverMountPath, `{"Name":"myvol","ID":"abc123"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response VolumeDriverMountResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Mountpoint != "/var/lib/docker-volumes/myvol" {
			t.Errorf("unexpected mountpoint: %q", response.Mountpoint)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			mountFn: func(_ VolumeDriverMountRequest) (*VolumeDriverMountResponse, error) {
				return nil, errors.New("mount failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverMountPath, `{"Name":"myvol","ID":"abc123"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverMountPath, `{invalid}`)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverUnmount(t *testing.T) {
	t.Run("happy path returns 200 with empty JSON object", func(t *testing.T) {
		recorder := doVolumeDriverRequest(&mockVolumeDriver{}, volumeDriverUnmountPath, `{"Name":"myvol","ID":"abc123"}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			unmountFn: func(_ VolumeDriverUnmountRequest) error {
				return errors.New("unmount failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverUnmountPath, `{"Name":"myvol","ID":"abc123"}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestVolumeDriverCapabilities(t *testing.T) {
	t.Run("happy path returns capabilities", func(t *testing.T) {
		driver := &mockVolumeDriver{
			capabilitiesFn: func() (*VolumeDriverCapabilitiesResponse, error) {
				return &VolumeDriverCapabilitiesResponse{
					Capabilities: VolumeDriverCapability{Scope: "global"},
				}, nil
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverCapabilitiesPath, `{}`)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		var response VolumeDriverCapabilitiesResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Capabilities.Scope != "global" {
			t.Errorf("expected scope %q, got %q", "global", response.Capabilities.Scope)
		}
	})

	t.Run("driver error returns 500 with Err field", func(t *testing.T) {
		driver := &mockVolumeDriver{
			capabilitiesFn: func() (*VolumeDriverCapabilitiesResponse, error) {
				return nil, errors.New("capabilities failed")
			},
		}

		recorder := doVolumeDriverRequest(driver, volumeDriverCapabilitiesPath, `{}`)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}
