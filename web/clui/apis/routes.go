/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/postgres"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

var UrlBefore string

func runArgs(cfg string) (args []interface{}) {
	host := "127.0.0.1"
	port := 443
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
	cert := viper.GetString("api.cert")
	key := viper.GetString("api.key")
	log.Printf("cert: %s, key: %s\n", cert, key)
	if cert != "" && key != "" {
		listen := viper.GetString("api.listen")
		http.ListenAndServeTLS(listen, cert, key, m)
	} else {
		m.Run(runArgs("api.listen")...)
	}
	return
}

func New() (m *macaron.Macaron) {
	m = macaron.Classic()
	m.Use(i18n.I18n(i18n.Options{
		Langs:       []string{"en-US", "zh-CN"},
		Names:       []string{"English", "简体中文"},
		DefaultLang: "zh-CN",
	}))
	m.Use(session.Sessioner(session.Options{
		IDLength:       16,
		Provider:       "postgres",
		ProviderConfig: viper.GetString("db.uri"),
	}))
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
	m.Post("/api/v1/login", userAPI.LoginPost)
	/*
	m.Get("/api/v1/hypers", hyperView.List)
	m.Get("/api/v1/users", userView.List)
	m.Get("/api/v1/users/:id", userView.Edit)
	m.Post("/api/v1/users/:id", userView.Patch)
	m.Get("/api/v1/users/:id/chorg", userView.Change)
	m.Delete("/api/v1/users/:id", userView.Delete)
	m.Get("/api/v1/users/new", userView.New)
	m.Post("/api/v1/users/new", userView.Create)
	m.Get("/api/v1/orgs", orgView.List)
	m.Get("/api/v1/orgs/:id", orgView.Edit)
	m.Post("/api/v1/orgs/:id", orgView.Patch)
	m.Delete("/api/v1/orgs/:id", orgView.Delete)
	m.Get("/api/v1/orgs/new", orgView.New)
	m.Post("/api/v1/orgs/new", orgView.Create)
	m.Get("/api/v1/instances", instanceView.List)
	m.Get("/api/v1/UpdateTable", instanceView.UpdateTable)
	m.Get("/api/v1/instances/new", instanceView.New)
	m.Post("/api/v1/instances/new", instanceView.Create)
	m.Delete("/api/v1/instances/:id", instanceView.Delete)
	m.Get("/api/v1/instances/:id", instanceView.Edit)
	m.Post("/api/v1/instances/:id", instanceView.Patch)
	m.Post("/api/v1/instances/:id/console", consoleView.ConsoleURL)
	m.Get("/api/v1/consoleresolver/token/:token", consoleView.ConsoleResolve)
	m.Get("/api/v1/interfaces/:id", interfaceView.Edit)
	m.Post("/api/v1/interfaces/:id", interfaceView.Patch)
	m.Post("/api/v1/interfaces/new", interfaceView.Create)
	m.Delete("/api/v1/interfaces/:id", interfaceView.Delete)
	m.Get("/api/v1/flavors", flavorView.List)
	m.Get("/api/v1/flavors/new", flavorView.New)
	m.Post("/api/v1/flavors/new", flavorView.Create)
	m.Delete("/api/v1/flavors/:id", flavorView.Delete)
	m.Get("/api/v1/images", imageView.List)
	m.Get("/api/v1/images/new", imageView.New)
	m.Post("/api/v1/images/new", imageView.Create)
	m.Delete("/api/v1/images/:id", imageView.Delete)
	m.Get("/api/v1/volumes", volumeView.List)
	m.Get("/api/v1/volumes/new", volumeView.New)
	m.Post("/api/v1/volumes/new", volumeView.Create)
	m.Delete("/api/v1/volumes/:id", volumeView.Delete)
	m.Get("/api/v1/volumes/:id", volumeView.Edit)
	m.Post("/api/v1/volumes/:id", volumeView.Patch)
	m.Get("/api/v1/subnets", subnetView.List)
	m.Get("/api/v1/subnets/new", subnetView.New)
	m.Post("/api/v1/subnets/new", subnetView.Create)
	m.Delete("/api/v1/subnets/:id", subnetView.Delete)
	m.Get("/api/v1/subnets/:id", subnetView.Edit)
	m.Post("/api/v1/subnets/:id", subnetView.Patch)
	m.Get("/api/v1/keys", keyView.List)
	m.Get("/api/v1/keys/new", keyView.New)
	m.Post("/api/v1/keys/new", keyView.Create)
	m.Post("/api/v1/keys/confirm", keyView.Confirm)
	m.Delete("/api/v1/keys/:id", keyView.Delete)
	m.Get("/api/v1/floatingips", floatingipView.List)
	m.Get("/api/v1/floatingips/new", floatingipView.New)
	m.Post("/api/v1/floatingips/new", floatingipView.Create)
	m.Post("/api/v1/floatingips/assign", floatingipView.Assign)
	m.Delete("/api/v1/floatingips/:id", floatingipView.Delete)
	m.Get("/api/v1/portmaps", portmapView.List)
	m.Get("/api/v1/portmaps/new", portmapView.New)
	m.Post("/api/v1/portmaps/new", portmapView.Create)
	m.Delete("/api/v1/portmaps/:id", portmapView.Delete)
	m.Get("/api/v1/routers", routerView.List)
	m.Get("/api/v1/routers/new", routerView.New)
	m.Post("/api/v1/routers/new", routerView.Create)
	m.Delete("/api/v1/routers/:id", routerView.Delete)
	m.Get("/api/v1/routers/:id", routerView.Edit)
	m.Post("/api/v1/routers/:id", routerView.Patch)
	m.Get("/api/v1/secgroups", secgroupView.List)
	m.Get("/api/v1/secgroups/new", secgroupView.New)
	m.Post("/api/v1/secgroups/new", secgroupView.Create)
	m.Delete("/api/v1/secgroups/:id", secgroupView.Delete)
	m.Get("/api/v1/secgroups/:id", secgroupView.Edit)
	m.Post("/api/v1/secgroups/:id", secgroupView.Patch)
	m.Get("/api/v1/secgroups/:sgid/secrules", secruleView.List)
	m.Get("/api/v1/secgroups/:sgid/secrules/new", secruleView.New)
	m.Post("/api/v1/secgroups/:sgid/secrules/new", secruleView.Create)
	m.Delete("/api/v1/secgroups/:sgid/secrules/:id", secruleView.Delete)
	m.Get("/api/v1/secgroups/:sgid/secrules/:id", secruleView.Edit)
	m.Post("/api/v1/secgroups/:sgid/secrules/:id", secruleView.Patch)
	*/
	m.NotFound(func(c *macaron.Context) {
		c.JSON(404, map[string]interface{}{
			"error":   "not found",
		})
	})
	return
}
