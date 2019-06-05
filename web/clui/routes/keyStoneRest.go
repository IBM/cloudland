/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"github.com/IBM/cloudland/web/clui/model"
	model "github.com/IBM/cloudland/web/rest-api/rest/models"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

var (
	versionInstance = &Versions{}
)

type Versions model.Versions

func (v *Versions) listVersion(c *macaron.Context, store session.Store) {
	v
}
