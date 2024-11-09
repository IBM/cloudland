/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package rpcs

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

func runArgs(cfg string) (args []interface{}) {
	host := "127.0.0.1"
	port := 50050
	listen := viper.GetString(cfg)
	if listen != "" {
		items := strings.Split(listen, ":")
		if len(items) == 2 {
			if items[0] != "" {
				host = items[0]
			}
			if items[1] != "" {
				port, _ = strconv.Atoi(items[1])
			}
		}
	}
	args = append(args, host, port)
	return
}

func Run() (err error) {
	m := New()
	m.Run(runArgs("internal.listen")...)
	return
}

func New() (m *macaron.Macaron) {
	m = macaron.Classic()
	m.Use(macaron.Renderer(
		macaron.RenderOptions{
			Funcs: []template.FuncMap{
				template.FuncMap{
					"GetString": viper.GetString,
					"Title":     func(v interface{}) string { return strings.Title(fmt.Sprint(v)) },
				},
			},
		},
	))
	m.Post("/internal/execute", frontbackService.Execute)
	return
}
