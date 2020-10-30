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
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/clui/model"
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
	m.Use(LinkHandler)
	m.Get("/", Index)
	m.Get("/dashboard", dashboard.Show)
	m.Get("/dashboard/getdata", dashboard.GetData)
	m.Get("/login", userView.LoginGet)
	m.Post("/login", userView.LoginPost)
	m.Get("/hypers", hyperView.List)
	m.Get("/users", userView.List)
	m.Get("/users/:id", userView.Edit)
	m.Post("/users/:id", userView.Patch)
	m.Get("/users/:id/chorg", userView.Change)
	m.Delete("/users/:id", userView.Delete)
	m.Get("/users/new", userView.New)
	m.Post("/users/new", userView.Create)
	m.Get("/orgs", orgView.List)
	m.Get("/orgs/:id", orgView.Edit)
	m.Post("/orgs/:id", orgView.Patch)
	m.Delete("/orgs/:id", orgView.Delete)
	m.Get("/orgs/new", orgView.New)
	m.Post("/orgs/new", orgView.Create)
	m.Get("/instances", instanceView.List)
	m.Get("/UpdateTable", instanceView.UpdateTable)
	m.Get("/instances/new", instanceView.New)
	m.Post("/instances/new", instanceView.Create)
	m.Delete("/instances/:id", instanceView.Delete)
	m.Get("/openshifts", openshiftView.List)
	m.Get("/openshifts/new", openshiftView.New)
	m.Post("/openshifts/new", openshiftView.Create)
	m.Delete("/openshifts/:id", openshiftView.Delete)
	m.Get("/openshifts/:id", openshiftView.Edit)
	m.Post("/openshifts/:id", openshiftView.Patch)
	m.Post("/openshifts/:id/launch", openshiftView.Launch)
	m.Post("/openshifts/:id/state", openshiftView.State)
	m.Get("/glusterfs", glusterfsView.List)
	m.Get("/glusterfs/new", glusterfsView.New)
	m.Post("/glusterfs/new", glusterfsView.Create)
	m.Delete("/glusterfs/:id", glusterfsView.Delete)
	m.Get("/glusterfs/:id", glusterfsView.Edit)
	m.Post("/glusterfs/:id", glusterfsView.Patch)
	m.Get("/instances/:id", instanceView.Edit)
	m.Post("/instances/:id", instanceView.Patch)
	m.Post("/instances/:id/console", consoleView.ConsoleURL)
	m.Get("/consoleresolver/token/:token", consoleView.ConsoleResolve)
	m.Get("/interfaces/:id", interfaceView.Edit)
	m.Post("/interfaces/:id", interfaceView.Patch)
	m.Post("/interfaces/new", interfaceView.Create)
	m.Delete("/interfaces/:id", interfaceView.Delete)
	m.Get("/flavors", flavorView.List)
	m.Get("/flavors/new", flavorView.New)
	m.Post("/flavors/new", flavorView.Create)
	m.Delete("/flavors/:id", flavorView.Delete)
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
	m.Get("/subnets/:id", subnetView.Edit)
	m.Post("/subnets/:id", subnetView.Patch)
	m.Get("/keys", keyView.List)
	m.Get("/keys/new", keyView.New)
	m.Post("/keys/new", keyView.Create)
	m.Post("/keys/confirm", keyView.Confirm)
	m.Delete("/keys/:id", keyView.Delete)
	m.Get("/floatingips", floatingipView.List)
	m.Get("/floatingips/new", floatingipView.New)
	m.Post("/floatingips/new", floatingipView.Create)
	m.Post("/floatingips/assign", floatingipView.Assign)
	m.Delete("/floatingips/:id", floatingipView.Delete)
	m.Get("/portmaps", portmapView.List)
	m.Get("/portmaps/new", portmapView.New)
	m.Post("/portmaps/new", portmapView.Create)
	m.Delete("/portmaps/:id", portmapView.Delete)
	m.Get("/gateways", gatewayView.List)
	m.Get("/gateways/new", gatewayView.New)
	m.Post("/gateways/new", gatewayView.Create)
	m.Delete("/gateways/:id", gatewayView.Delete)
	m.Get("/gateways/:id", gatewayView.Edit)
	m.Post("/gateways/:id", gatewayView.Patch)
	m.Get("/secgroups", secgroupView.List)
	m.Get("/secgroups/new", secgroupView.New)
	m.Post("/secgroups/new", secgroupView.Create)
	m.Delete("/secgroups/:id", secgroupView.Delete)
	m.Get("/secgroups/:id", secgroupView.Edit)
	m.Post("/secgroups/:id", secgroupView.Patch)
	m.Get("/secgroups/:sgid/secrules", secruleView.List)
	m.Get("/secgroups/:sgid/secrules/new", secruleView.New)
	m.Post("/secgroups/:sgid/secrules/new", secruleView.Create)
	m.Delete("/secgroups/:sgid/secrules/:id", secruleView.Delete)
	m.Get("/error", func(c *macaron.Context) {
		c.Data["ErrorMsg"] = c.QueryTrim("ErrorMsg")
		c.HTML(500, "error")
	})
	m.NotFound(func(c *macaron.Context) { c.HTML(404, "404") })
	return
}

func LinkHandler(c *macaron.Context, store session.Store) {
	UrlBefore = "/"
	link := strings.NewReplacer("%", "%25",
		"#", "%23",
		" ", "%20",
		"?", "%3F").Replace(
		strings.TrimSuffix(c.Req.URL.Path, "/"))
	log.Println(link)
	c.Data["Link"] = link
	if login, ok := store.Get("login").(string); ok {
		// log.Println("$$$$$$$$$$$$$$$$$$", c.Locale.Language())
		memberShip := &MemberShip{
			OrgID:    store.Get("oid").(int64),
			UserID:   store.Get("uid").(int64),
			UserName: store.Get("login").(string),
			OrgName:  store.Get("org").(string),
			Role:     store.Get("role").(model.Role),
		}
		c.Req.Request = c.Req.WithContext(memberShip.SetContext(c.Req.Context()))
		c.Data["IsSignedIn"] = true
		if memberShip.Role == model.Admin || login == "admin" {
			c.Data["IsAdmin"] = true
			memberShip.Role = model.Admin
		}
		c.Data["Organization"] = store.Get("org").(string)
		c.Data["Members"] = store.Get("members").([]*model.Member)
	} else if link != "" && link != "/" && !strings.HasPrefix(link, "/login") && !strings.HasPrefix(link, "/consoleresolver") {
		UrlBefore=link
		c.Redirect("login?redirect_to=")
	}
}
