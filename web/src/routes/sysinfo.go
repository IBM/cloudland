/*
Copyright <holder> All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package routes

import "os"

var (
	Version      = "unknown"
	sysInfoAdmin = &SysInfoAdmin{}
)

type SysInfoAdmin struct{}

const (
	VersionFile = "/opt/cloudland/version"
)

func (v *SysInfoAdmin) GetVersion() string {
	if Version == "unknown" {
		// read version from file
		// check if file exists
		// if not, return default version
		// if yes, read version from file
		// return version
		version, err := os.ReadFile(VersionFile)
		if err != nil {
			logger.Warningf("failed to read version file: %v", err)
		} else {
			Version = string(version)
		}
	}
	return Version
}
