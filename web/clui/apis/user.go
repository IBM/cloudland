/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	macaron "gopkg.in/macaron.v1"
)

var userAPI = &UserAPI{}

type UserAPI struct{}

func (v *UserAPI) LoginPost(c *macaron.Context) {
	c.JSON(200, map[string]interface{}{
		"user":   "test",
	})
}

