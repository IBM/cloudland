/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package routes

import (
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

func Index(c *macaron.Context, store session.Store) {
	logout := c.Query("logout")
	if logout != "" {
		redirectTo := ""
		store.Destory(c)
		c.Redirect(redirectTo)
	} else {
		c.HTML(200, "index")
	}
}

func Admin(c *macaron.Context) {
	c.HTML(200, "dashboard")
}
