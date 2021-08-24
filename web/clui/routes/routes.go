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
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cors"
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
	m.Use(cors.CORS())
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
	m.Post("/api/login", binding.Bind(APIUserView{}), apiUserView.LoginPost)
	m.Get("/hypers", hyperView.List)
	m.Get("/api/hypers", apiHyperView.List)
	m.Get("/users", userView.List)
	m.Get("/api/users", apiUserView.List)
	m.Get("/users/:id", userView.Edit)
	m.Get("/api/users/:id", apiUserView.Edit)
	m.Post("/users/:id", userView.Patch)
	m.Post("/api/users/:id", binding.Bind(APIUserView{}), apiUserView.Patch)
	m.Get("/users/:id/chorg", userView.Change)
	m.Delete("/users/:id", userView.Delete)
	m.Delete("/api/users/:id", apiUserView.Delete)
	m.Get("/users/new", userView.New)
	m.Post("/users/new", userView.Create)
	m.Post("/api/users/new", binding.Bind(APIUserView{}), apiUserView.Create)
	m.Get("/orgs", orgView.List)
	m.Get("/api/orgs", apiOrgView.List)
	m.Get("/orgs/:id", orgView.Edit)
	m.Get("/api/orgs/:id", apiOrgView.Edit)
	m.Post("/orgs/:id", orgView.Patch)
	m.Post("/api/orgs/:id", binding.Bind(APIOrgView{}), apiOrgView.Patch)
	m.Delete("/orgs/:id", orgView.Delete)
	m.Delete("/api/orgs/:id", orgView.Delete)
	m.Get("/orgs/new", orgView.New)
	m.Post("/orgs/new", orgView.Create)
	m.Post("/api/orgs/new", binding.Bind(APIOrgView{}), apiOrgView.Create)
	m.Get("/instances", instanceView.List)
	m.Get("/api/instances", apiInstanceView.List)
	m.Get("/UpdateTable", instanceView.UpdateTable)
	m.Get("/instances/new", instanceView.New)
	m.Post("/instances/new", instanceView.Create)
	m.Post("/api/instances/new", binding.Bind(APIInstanceView{}), apiInstanceView.Create)
	m.Delete("/instances/:id", instanceView.Delete)
	m.Delete("/api/instances/:id", apiInstanceView.Delete)
	m.Get("/openshifts", openshiftView.List)
	m.Get("/api/openshifts", apiOpenshiftView.List)
	m.Get("/openshifts/new", openshiftView.New)
	m.Post("/openshifts/new", openshiftView.Create)
	m.Post("/api/openshifts/new", binding.Bind(APIOpenshiftView{}), apiOpenshiftView.Create)
	m.Delete("/openshifts/:id", openshiftView.Delete)
	m.Delete("/api/openshifts/:id", apiOpenshiftView.Delete)
	m.Get("/openshifts/:id", openshiftView.Edit)
	m.Get("/api/openshifts/:id", apiOpenshiftView.Edit)
	m.Post("/openshifts/:id", openshiftView.Patch)
	m.Post("/api/openshifts/:id", binding.Bind(APIOpenshiftView{}), apiOpenshiftView.Patch)
	m.Post("/openshifts/:id/launch", openshiftView.Launch)
	m.Post("/api/openshifts/:id/launch", binding.Bind(APIOpenshiftView{}), apiOpenshiftView.Launch)
	m.Post("/openshifts/:id/state", openshiftView.State)
	m.Post("/api/openshifts/:id/state", binding.Bind(APIOpenshiftView{}), apiOpenshiftView.State)
	m.Get("/glusterfs", glusterfsView.List)
	m.Get("/api/glusterfs", apiGlusterfsView.List)
	m.Get("/glusterfs/new", glusterfsView.New)
	m.Post("/glusterfs/new", glusterfsView.Create)
	m.Post("/api/glusterfs/new", binding.Bind(APIGlusterfsView{}), apiGlusterfsView.Create)
	m.Delete("/glusterfs/:id", glusterfsView.Delete)
	m.Delete("/api/glusterfs/:id", apiGlusterfsView.Delete)
	m.Get("/glusterfs/:id", glusterfsView.Edit)
	m.Get("/api/glusterfs/:id", apiGlusterfsView.Edit)
	m.Post("/glusterfs/:id", glusterfsView.Patch)
	m.Post("/api/glusterfs/:id", binding.Bind(APIGlusterfsView{}), apiGlusterfsView.Patch)
	m.Get("/instances/:id", instanceView.Edit)
	m.Get("/api/instances/:id", apiInstanceView.Edit)
	m.Post("/instances/:id", instanceView.Patch)
	m.Post("/api/instances/:id", binding.Bind(APIInstanceView{}), apiInstanceView.Patch)
	m.Post("/instances/:id/console", consoleView.ConsoleURL)
	m.Post("/api/instances/:id/console", apiConsoleView.ConsoleURL)
	m.Get("/consoleresolver/token/:token", consoleView.ConsoleResolve)
	m.Get("/api/consoleresolver/token/:token", apiConsoleView.ConsoleResolve)
	m.Get("/interfaces/:id", interfaceView.Edit)
	m.Get("/api/interfaces/:id", apiInterfaceView.Edit)
	m.Post("/interfaces/:id", interfaceView.Patch)
	m.Post("/api/interfaces/:id", binding.Bind(APIInterfaceView{}), apiInterfaceView.Patch)
	m.Post("/interfaces/new", interfaceView.Create)
	m.Post("/api/interfaces/new", binding.Bind(APIInterfaceView{}), apiInterfaceView.Create)
	m.Delete("/interfaces/:id", interfaceView.Delete)
	m.Delete("/api/interfaces/:id", apiInterfaceView.Delete)
	m.Get("/flavors", flavorView.List)
	m.Get("/api/flavors", apiFlavorView.List)
	m.Get("/flavors/new", flavorView.New)
	m.Post("/flavors/new", flavorView.Create)
	m.Post("/api/flavors/new", binding.Bind(APIFlavorView{}), apiFlavorView.Create)
	m.Delete("/flavors/:id", flavorView.Delete)
	m.Delete("/api/flavors/:id", apiFlavorView.Delete)
	m.Get("/registrys", registryView.List)
	m.Get("/api/registrys", apiRegistryView.List)
	m.Get("/registrys/new", registryView.New)
	m.Post("/registrys/new", registryView.Create)
	m.Post("/api/registrys/new", binding.Bind(APIRegistryView{}), apiRegistryView.Create)
	m.Delete("/registrys/:id", registryView.Delete)
	m.Delete("/api/registrys/:id", apiRegistryView.Delete)
	m.Get("/registrys/:id", registryView.Edit)
	m.Get("/api/registrys/:id", apiRegistryView.Edit)
	m.Post("/registrys/:id", registryView.Patch)
	m.Post("/api/registrys/:id", binding.Bind(APIRegistryView{}), apiRegistryView.Patch)
	m.Get("/images", imageView.List)
	m.Get("/api/images", apiImageView.List)
	m.Get("/images/new", imageView.New)
	m.Post("/images/new", imageView.Create)
	m.Post("/api/images/new", binding.Bind(APIImageView{}), apiImageView.Create)
	m.Delete("/images/:id", imageView.Delete)
	m.Delete("/api/images/:id", apiImageView.Delete)
	m.Get("/volumes", volumeView.List)
	m.Get("/api/volumes", apiVolumeView.List)
	m.Get("/volumes/new", volumeView.New)
	m.Post("/volumes/new", volumeView.Create)
	m.Post("/api/volumes/new", binding.Bind(APIVolumeView{}), apiVolumeView.Create)
	m.Delete("/volumes/:id", volumeView.Delete)
	m.Delete("/api/volumes/:id", apiVolumeView.Delete)
	m.Get("/volumes/:id", volumeView.Edit)
	m.Get("/api/volumes/:id", apiVolumeView.Edit)
	m.Post("/volumes/:id", volumeView.Patch)
	m.Post("/api/volumes/:id", binding.Bind(APIVolumeView{}), apiVolumeView.Patch)
	m.Get("/subnets", subnetView.List)
	m.Get("/api/subnets", apiSubnetView.List)
	m.Get("/subnets/new", subnetView.New)
	m.Post("/subnets/new", subnetView.Create)
	m.Post("/api/subnets/new", binding.Bind(APISubnetView{}), apiSubnetView.Create)
	m.Delete("/subnets/:id", subnetView.Delete)
	m.Delete("/api/subnets/:id", apiSubnetView.Delete)
	m.Get("/subnets/:id", subnetView.Edit)
	m.Get("/api/subnets/:id", apiSubnetView.Query)
	m.Post("/subnets/:id", subnetView.Patch)
	m.Post("/api/subnets/:id", binding.Bind(APISubnetView{}), apiSubnetView.Patch)
	m.Get("/keys", keyView.List)
	m.Get("/api/keys", apiKeyView.List)
	m.Get("/keys/new", keyView.New)
	m.Post("/keys/new", keyView.Create)
	m.Post("/api/keys/new", binding.Bind(APIKeyView{}), apiKeyView.Create)
	m.Post("/keys/confirm", keyView.Confirm)
	m.Delete("/keys/:id", keyView.Delete)
	m.Delete("/api/keys/:id", apiKeyView.Delete)
	m.Delete("/api/keys/:id", apiKeyView.Delete)
	m.Get("/floatingips", floatingipView.List)
	m.Get("/api/floatingips", apiFloatingipView.List)
	m.Get("/floatingips/new", floatingipView.New)
	m.Post("/floatingips/new", floatingipView.Create)
	m.Post("/api/floatingips/new", binding.Bind(APIFloatingIpView{}), apiFloatingipView.Create)
	m.Post("/floatingips/assign", floatingipView.Assign)
	m.Post("/api/floatingips/assign", binding.Bind(APIFloatingIpView{}), apiFloatingipView.Assign)
	m.Delete("/api/floatingips/:id", floatingipView.Delete)
	m.Delete("/api/floatingips/:id", apiFloatingipView.Delete)
	m.Get("/portmaps", portmapView.List)
	m.Get("/api/portmaps", apiPortmapView.List)
	m.Get("/portmaps/new", portmapView.New)
	m.Post("/portmaps/new", portmapView.Create)
	m.Post("/api/portmaps/new", binding.Bind(APIPortmapView{}), apiPortmapView.Create)
	m.Delete("/portmaps/:id", portmapView.Delete)
	m.Delete("/api/portmaps/:id", apiPortmapView.Delete)
	m.Get("/gateways", gatewayView.List)
	m.Get("/api/gateways", apiGatewayView.List)
	m.Get("/gateways/new", gatewayView.New)
	m.Post("/gateways/new", gatewayView.Create)
	m.Post("/api/gateways/new", binding.Bind(APIGatewayView{}), apiGatewayView.Create)
	m.Delete("/gateways/:id", gatewayView.Delete)
	m.Delete("/api/gateways/:id", apiGatewayView.Delete)
	m.Get("/gateways/:id", gatewayView.Edit)
	m.Get("/api/gateways/:id", apiGatewayView.Edit)
	m.Post("/gateways/:id", gatewayView.Patch)
	m.Post("/api/gateways/:id", binding.Bind(APIGatewayPatch{}), apiGatewayPatch.Patch)
	m.Get("/secgroups", secgroupView.List)
	m.Get("/api/secgroups", apiSecgroupView.List)
	m.Get("/secgroups/new", secgroupView.New)
	m.Post("/secgroups/new", secgroupView.Create)
	m.Post("/api/secgroups/new", binding.Bind(APISecgroupView{}), apiSecgroupView.Create)
	m.Delete("/secgroups/:id", secgroupView.Delete)
	m.Delete("/api/secgroups/:id", apiSecgroupView.Delete)
	m.Get("/secgroups/:id", secgroupView.Edit)
	m.Get("/api/secgroups/:id", apiSecgroupView.Edit)
	m.Post("/secgroups/:id", secgroupView.Patch)
	m.Post("/api/secgroups/:id", binding.Bind(APISecgroupView{}), apiSecgroupView.Patch)
	m.Get("/secgroups/:sgid/secrules", secruleView.List)
	m.Get("/api/secgroups/:sgid/secrules", apiSecruleView.List)
	m.Get("/secgroups/:sgid/secrules/new", secruleView.New)
	m.Post("/secgroups/:sgid/secrules/new", secruleView.Create)
	m.Post("/api/secgroups/:sgid/secrules/new", binding.Bind(APISecruleView{}), apiSecruleView.Create)
	m.Delete("/secgroups/:sgid/secrules/:id", secruleView.Delete)
	m.Delete("/api/secgroups/:sgid/secrules/:id", apiSecruleView.Delete)
	m.Get("/secgroups/:sgid/secrules/:id", secruleView.Edit)
	m.Get("/api/secgroups/:sgid/secrules/:id", apiSecruleView.Edit)
	m.Post("/secgroups/:sgid/secrules/:id", secruleView.Patch)
	m.Post("/api/secgroups/:sgid/secrules/:id", binding.Bind(APISecruleView{}), apiSecruleView.Patch)
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
	if link != "/UpdateTable" {
		log.Println(link)
	}
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
	} else if link != "" && link != "/" && !strings.HasPrefix(link, "/login") && !strings.HasPrefix(link, "/consoleresolver") && strings.HasPrefix(link, "/api") {

		if strings.HasPrefix(link, "/api") {
			if !strings.HasPrefix(link, "/api/login") {
				//parse token
				token := c.Req.Header.Get("X-Auth-Token")
				claims, err := ParseToken(token)
				if err != nil {
					log.Println(err.Error())
					c.JSON(401, map[string]interface{}{
						"ErrorMsg": "token unauthorized",
					})
					return
				} else {
					//uid := claims.UID
					//oid := claims.OID
					username := claims.StandardClaims.Audience
					organization := username
					org, err := orgAdmin.Get(organization)
					role := claims.Role
					members := []*model.Member{}
					err = DB().Where("user_name = ?", username).Find(&members).Error
					if err != nil {
						log.Println("Failed to query organizations, ", err)
						c.JSON(403, map[string]interface{}{
							"ErrorMsg": "Failed to query organizations",
						})
						return
					}

					user, err := userAdmin.Get(username)
					if err != nil {
						log.Println("Failed to query user, ", err)
						c.JSON(403, map[string]interface{}{
							"ErrorMsg": "Failed to query user",
						})
						return
					}

					//set store
					store.Set("login", username)
					store.Set("uid", user.ID)
					store.Set("oid", org.ID)
					store.Set("role", role)
					store.Set("act", token)
					store.Set("org", organization)
					store.Set("defsg", org.DefaultSG)
					store.Set("members", members)
					//set membership to request context
					memberShip := &MemberShip{
						OrgID:    store.Get("oid").(int64),
						UserID:   store.Get("uid").(int64),
						UserName: store.Get("login").(string),
						OrgName:  store.Get("org").(string),
						Role:     store.Get("role").(model.Role),
					}
					c.Req.Request = c.Req.WithContext(memberShip.SetContext(c.Req.Context()))

					if memberShip.Role == model.Admin || username == "admin" {
						memberShip.Role = model.Admin
					}

				}

			}

		} else {
			UrlBefore = link
			c.Redirect("login?redirect_to=")
		}

	}

}
