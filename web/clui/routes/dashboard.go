/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

func Dashboard(c *macaron.Context, store session.Store) {
	c.HTML(200, "dashboard")
}
