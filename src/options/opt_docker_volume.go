// SPDX-FileCopyrightText: © 2024 - 2025 Anthony Champagne <dev@anthonychampagne.fr>
//
// SPDX-License-Identifier: AGPL-3.0-only

package options

import (
	"fmt"
	"strconv"
)

type OptDockerVolume struct {
	MountUId       uint16 `json:","`
	MountGId       uint16 `json:","`
	MountMode      uint32 `json:","`
	FieldMountMode uint32 `json:","`
}

func (z OptDockerVolume) CacheId_() string {
	r := ""

	r += strconv.Itoa(int(z.MountUId)) +
		strconv.Itoa(int(z.MountGId)) +
		strconv.Itoa(int(z.MountMode)) +
		strconv.Itoa(int(z.FieldMountMode))

	return r
}

func MakeOptDockerVolume() OptDockerVolume {
	return OptDockerVolume{
		MountMode:      0550,
		FieldMountMode: 0440,
	}
}

func NewOptDockerVolume(volumeName string, volumeOptions map[string]string, defaultConfig *OptDockerVolume) (*OptDockerVolume, error) {
	var r OptDockerVolume

	if defaultConfig != nil {
		r = *defaultConfig
	}

	if err := r.Update(volumeName, volumeOptions); err != nil {
		return nil, err
	}

	if err := r.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (z *OptDockerVolume) Update(volumeName string, volumeOptions map[string]string) error {
	vmui, ok := volumeOptions["mount-uid"]
	if ok {
		v, err := strconv.Atoi(vmui)
		if err != nil {
			return fmt.Errorf("convert mount-uid %s to integer: %w", vmui, err)
		}

		z.MountUId = uint16(v)
	}

	vmmg, ok := volumeOptions["mount-gid"]
	if ok {
		v, err := strconv.Atoi(vmmg)
		if err != nil {
			return fmt.Errorf("convert mount-gid %s to integer: %w", vmmg, err)
		}

		z.MountGId = uint16(v)
	}

	vomm, ok := volumeOptions["mount-mode"]
	if ok {
		v, err := strconv.Atoi(vomm)
		if err != nil {
			return fmt.Errorf("convert mount-mode %s to integer: %w", vomm, err)
		}

		z.MountMode = uint32(v)
	}

	vofmm, ok := volumeOptions["field-mount-mode"]
	if ok {
		v, err := strconv.Atoi(vofmm)
		if err != nil {
			return fmt.Errorf("convert mount-uid %s to integer: %w", vofmm, err)
		}

		z.FieldMountMode = uint32(v)
	}

	return nil
}

func (z *OptDockerVolume) Normalize() {}

func (z *OptDockerVolume) NormalizeAndValidate() error {
	z.Normalize()

	return nil
}
