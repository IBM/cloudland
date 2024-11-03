/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"github.com/IBM/cloudland/web/src/routes"
)

var routerAPI = &RouterAPI{}
var routerAdmin = &routes.RouterAdmin{}

type RouterAPI struct{}

type RouterInfo struct {
}
