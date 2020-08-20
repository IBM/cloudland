/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

var (
	consoleAdmin = &ConsoleAdmin{}
	consoleView  = &ConsoleView{}
)

type ConsoleAdmin struct{}
type ConsoleView struct{}

type ConsoleInfo struct{}

func (a *ConsoleView) ConsoleURL(c *macaron.Context, store session.Store) {
	consoleURL := "http://cloudland.bluecat.ltd"
	c.Resp.Header().Set("Location", consoleURL)
	c.JSON(301, nil)
	return
}

func (a *ConsoleView) ConsoleResolve(c *macaron.Context, store session.Store) {
	consoleInfo := &ConsoleInfo{}
	c.JSON(200, consoleInfo)
	return
}
