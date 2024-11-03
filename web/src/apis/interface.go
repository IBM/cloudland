/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"github.com/IBM/cloudland/web/src/routes"
)

var interfaceAPI = &InterfaceAPI{}
var interfaceAdmin = &routes.InterfaceAdmin{}

type InterfaceAPI struct{}

type InterfaceInfo struct {
	*BaseReference
	MacAddress  string            `json:"mac_address"`
	IPAddress   string            `json:"ip_address"`
	IsPrimary bool           `json:"is_primary"`
	FloatingIps []*FloatingIpInfo `json:"floating_ips,omitempty"`
}
