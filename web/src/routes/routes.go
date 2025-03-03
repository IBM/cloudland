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
	"net/http"
	"strconv"
	"strings"

	. "web/src/common"
	"web/src/model"
	rlog "web/src/utils/log"

	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/postgres"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

var UrlBefore string
var logger = rlog.MustGetLogger("routes")

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
	logger.Info("Start to run cloudland base service")
	m := New()
	cert := viper.GetString("base.cert")
	key := viper.GetString("base.key")
	logger.Debugf("cert: %s, key: %s\n", cert, key)
	listen := viper.GetString("base.listen")
	if cert != "" && key != "" {
		logger.Infof("Running https service isten on %s\n", listen)
		http.ListenAndServeTLS(listen, cert, key, m)
	} else {
		logger.Infof("Running http service on %s\n", listen)
		m.Run(runArgs("base.listen")...)
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
	adminInit()
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
	m.Use(SysVersion)
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
	m.Get("/instances/:id", instanceView.Status)
	m.Get("/UpdateTable", instanceView.UpdateTable)
	m.Get("/instances/new", instanceView.New)
	m.Post("/instances/new", instanceView.Create)
	m.Delete("/instances/:id", instanceView.Delete)
	m.Get("/instances/:id", instanceView.Edit)
	m.Post("/instances/:id", instanceView.Patch)
	m.Get("/instances/:id/set_user_password", instanceView.SetUserPassword)
	m.Post("/instances/:id/set_user_password", instanceView.SetUserPassword)
	m.Post("/instances/:id/console", consoleView.ConsoleURL)
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
	m.Get("/floatingips", floatingIpView.List)
	m.Get("/floatingips/new", floatingIpView.New)
	m.Post("/floatingips/new", floatingIpView.Create)
	m.Delete("/floatingips/:id", floatingIpView.Delete)
	m.Get("/portmaps", portmapView.List)
	m.Get("/portmaps/new", portmapView.New)
	m.Post("/portmaps/new", portmapView.Create)
	m.Delete("/portmaps/:id", portmapView.Delete)
	m.Get("/routers", routerView.List)
	m.Get("/routers/new", routerView.New)
	m.Post("/routers/new", routerView.Create)
	m.Delete("/routers/:id", routerView.Delete)
	m.Get("/routers/:id", routerView.Edit)
	m.Post("/routers/:id", routerView.Patch)
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
	m.Get("/secgroups/:sgid/secrules/:id", secruleView.Edit)
	m.Post("/secgroups/:sgid/secrules/:id", secruleView.Patch)
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
	logger.Debugf("LinkHandler: %s\n", link)
	c.Data["Link"] = link
	if login, ok := store.Get("login").(string); ok {
		// logger.Debug("$$$$$$$$$$$$$$$$$$", c.Locale.Language())
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
	} else if link != "" && link != "/" && !strings.HasPrefix(link, "/login") {
		UrlBefore = link
		c.Redirect("login?redirect_to=")
	}
}

func SysVersion(c *macaron.Context, store session.Store) {
	c.Data["Version"] = sysInfoAdmin.GetVersion()
}
