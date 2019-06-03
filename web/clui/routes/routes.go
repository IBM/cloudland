/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

func runArgs() (args []interface{}) {
	host := "127.0.0.1"
	port := "4000"
	listen := viper.GetString("api.listen")
	if listen != "" {
		items := strings.Split(listen, ":")
		if len(items) == 2 {
			if items[0] != "" {
				host = items[0]
			}
			if items[1] != "" {
				port = items[1]
			}
		}
	}
	args = append(args, host, port)
	return
}

func Run() (err error) {
	New().Run(runArgs()...)
	return
}

func New() (m *macaron.Macaron) {
	m = macaron.Classic()
	m.Use(session.Sessioner())
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
	m.Use(LinkHandler)
	m.Get("/", Index)
	m.Get("/dashboard", Dashboard)
	m.Get("/login", userView.LoginGet)
	m.Post("/login", userView.LoginPost)
	m.Get("/users", userView.List)
	m.Get("/users/:id", userView.Show)
	m.Post("/users/:id", userView.Update)
	m.Delete("/users/:id", userView.Delete)
	m.Get("/users/new", userView.New)
	m.Post("/users/new", userView.Create)
	m.Get("/orgs", orgView.List)
	//	m.Get("/orgs/:id", orgView.Show)
	//	m.Post("/orgs/:id", orgView.Update)
	m.Delete("/orgs/:id", orgView.Delete)
	m.Get("/orgs/new", orgView.New)
	m.Post("/orgs/new", orgView.Create)
	m.Get("/instances", instanceView.List)
	m.Get("/instances/new", instanceView.New)
	m.Post("/instances/new", instanceView.Create)
	m.Delete("/instances/:id", instanceView.Delete)
	m.Get("/flavors", flavorView.List)
	m.Get("/flavors", flavorView.List)
	m.Get("/flavors/new", flavorView.New)
	m.Post("/flavors/new", flavorView.Create)
	m.Get("/images", imageView.List)
	m.Get("/images/new", imageView.New)
	m.Post("/images/new", imageView.Create)
	m.Delete("/images/:id", imageView.Delete)
	m.Get("/volumes", volumeView.List)
	m.Get("/volumes/new", volumeView.New)
	m.Post("/volumes/new", volumeView.Create)
	m.Delete("/volumes/:id", volumeView.Delete)
	m.Get("/volumes/:id", volumeView.Edit)
	m.Post("/volumes/:id", volumeView.Patch)
	m.Get("/subnets", subnetView.List)
	m.Get("/subnets/new", subnetView.New)
	m.Post("/subnets/new", subnetView.Create)
	m.Delete("/subnets/:id", subnetView.Delete)
	m.Get("/keys", keyView.List)
	m.Get("/keys/new", keyView.New)
	m.Post("/keys/new", keyView.Create)
	m.Delete("/keys/:id", keyView.Delete)
	m.Get("/floatingips", floatingipView.List)
	m.Get("/floatingips/new", floatingipView.New)
	m.Post("/floatingips/new", floatingipView.Create)
	m.Delete("/floatingips/:id", floatingipView.Delete)
	m.Get("/portmaps", portmapView.List)
	m.Get("/portmaps/new", portmapView.New)
	m.Post("/portmaps/new", portmapView.Create)
	m.Delete("/portmaps/:id", portmapView.Delete)
	m.Get("/gateways", gatewayView.List)
	m.Get("/gateways/new", gatewayView.New)
	m.Post("/gateways/new", gatewayView.Create)
	m.Delete("/gateways/:id", gatewayView.Delete)
	m.Get("/secgroups", secgroupView.List)
	m.Get("/secgroups/new", secgroupView.New)
	m.Post("/secgroups/new", secgroupView.Create)
	m.Delete("/secgroups/:id", secgroupView.Delete)
	m.Get("/secgroups/:sgid/secrules", secruleView.List)
	m.Get("/secgroups/:sgid/secrules/new", secruleView.New)
	m.Post("/secgroups/:sgid/secrules/new", secruleView.Create)
	m.Delete("/secgroups/:sgid/secrules/:id", secruleView.Delete)
	m.NotFound(func(c *macaron.Context) { c.HTML(404, "404") })
	return
}

func LinkHandler(c *macaron.Context, store session.Store) {
	link := strings.NewReplacer("%", "%25",
		"#", "%23",
		" ", "%20",
		"?", "%3F").Replace(
		strings.TrimSuffix(c.Req.URL.Path, "/"))
	log.Println(link)
	c.Data["Link"] = link
	if login, ok := store.Get("login").(string); ok {
		c.Data["IsSignedIn"] = true
		if login == "admin" {
			c.Data["IsAdmin"] = true
		}
		c.Data["Organization"] = store.Get("org").(string)
	} else if link != "" && link != "/" && !strings.HasPrefix(link, "/login") {
		c.Redirect("/")
	}
}
