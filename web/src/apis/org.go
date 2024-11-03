/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"github.com/IBM/cloudland/web/src/routes"
)

var orgAPI = &OrgAPI{}
var orgAdmin = &routes.OrgAdmin{}

type OrgAPI struct{}

type Organization struct {
	Name string `json:"name,required"`
	ID   string `json:"id,omitempty"`
}
