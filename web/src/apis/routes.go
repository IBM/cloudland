/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"
	"os"

	_ "github.com/IBM/cloudland/web/src/docs"
	"github.com/gin-gonic/gin"
	_ "github.com/go-macaron/session/postgres"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() (err error) {
	r := Register()
	cert := viper.GetString("rest.cert")
	key := viper.GetString("rest.key")
	listen := viper.GetString("rest.listen")
	log.Printf("cert: %s, key: %s\n", cert, key)
	if cert != "" && key != "" {
		r.RunTLS(listen, cert, key)
	} else {
		r.Run(listen)
	}
	return
}

// @title CloudLand API
// @version 1.0
// @description APIs for CloudLand Functions
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1
func Register() (r *gin.Engine) {
	r = gin.New()
	f, err := os.Create("/opt/cloudland/clapi.log")
	if err != nil {
		r.Use(gin.LoggerWithWriter(f))
	}
	r.Use(gin.Recovery())

	r.POST("/api/v1/login", userAPI.LoginPost)
	/*
		r.Get("/api/v1/hypers", hyperAPI.List)
		r.Get("/api/v1/users", userAPI.List)
		r.Get("/api/v1/users/:id", userAPI.Edit)
		r.Post("/api/v1/users/:id", userAPI.Patch)
		r.Get("/api/v1/users/:id/chorg", userAPI.Change)
		r.Delete("/api/v1/users/:id", userAPI.Delete)
		r.Get("/api/v1/users/new", userAPI.New)
		r.Post("/api/v1/users/new", userAPI.Create)
		r.Get("/api/v1/orgs", orgAPI.List)
		r.Get("/api/v1/orgs/:id", orgAPI.Edit)
		r.Post("/api/v1/orgs/:id", orgAPI.Patch)
		r.Delete("/api/v1/orgs/:id", orgAPI.Delete)
		r.Get("/api/v1/orgs/new", orgAPI.New)
		r.Post("/api/v1/orgs/new", orgAPI.Create)
		r.Get("/api/v1/instances", instanceAPI.List)
		r.Get("/api/v1/UpdateTable", instanceAPI.UpdateTable)
		r.Get("/api/v1/instances/new", instanceAPI.New)
		r.Post("/api/v1/instances/new", instanceAPI.Create)
		r.Delete("/api/v1/instances/:id", instanceAPI.Delete)
		r.Get("/api/v1/instances/:id", instanceAPI.Edit)
		r.Post("/api/v1/instances/:id", instanceAPI.Patch)
		r.Post("/api/v1/instances/:id/console", consoleAPI.ConsoleURL)
		r.Get("/api/v1/consoleresolver/token/:token", consoleAPI.ConsoleResolve)
		r.Get("/api/v1/interfaces/:id", interfaceAPI.Edit)
		r.Post("/api/v1/interfaces/:id", interfaceAPI.Patch)
		r.Post("/api/v1/interfaces/new", interfaceAPI.Create)
		r.Delete("/api/v1/interfaces/:id", interfaceAPI.Delete)
		r.Get("/api/v1/flavors", flavorAPI.List)
		r.Get("/api/v1/flavors/new", flavorAPI.New)
		r.Post("/api/v1/flavors/new", flavorAPI.Create)
		r.Delete("/api/v1/flavors/:id", flavorAPI.Delete)
		r.Get("/api/v1/images", imageAPI.List)
		r.Get("/api/v1/images/new", imageAPI.New)
		r.Post("/api/v1/images/new", imageAPI.Create)
		r.Delete("/api/v1/images/:id", imageAPI.Delete)
		r.Get("/api/v1/volumes", volumeAPI.List)
		r.Get("/api/v1/volumes/new", volumeAPI.New)
		r.Post("/api/v1/volumes/new", volumeAPI.Create)
		r.Delete("/api/v1/volumes/:id", volumeAPI.Delete)
		r.Get("/api/v1/volumes/:id", volumeAPI.Edit)
		r.Post("/api/v1/volumes/:id", volumeAPI.Patch)
		r.Get("/api/v1/subnets", subnetAPI.List)
		r.Get("/api/v1/subnets/new", subnetAPI.New)
		r.Post("/api/v1/subnets/new", subnetAPI.Create)
		r.Delete("/api/v1/subnets/:id", subnetAPI.Delete)
		r.Get("/api/v1/subnets/:id", subnetAPI.Edit)
		r.Post("/api/v1/subnets/:id", subnetAPI.Patch)
		r.Get("/api/v1/keys", keyAPI.List)
		r.Get("/api/v1/keys/new", keyAPI.New)
		r.Post("/api/v1/keys/new", keyAPI.Create)
		r.Post("/api/v1/keys/confirm", keyAPI.Confirm)
		r.Delete("/api/v1/keys/:id", keyAPI.Delete)
		r.Get("/api/v1/floatingips", floatingipAPI.List)
		r.Get("/api/v1/floatingips/new", floatingipAPI.New)
		r.Post("/api/v1/floatingips/new", floatingipAPI.Create)
		r.Post("/api/v1/floatingips/assign", floatingipAPI.Assign)
		r.Delete("/api/v1/floatingips/:id", floatingipAPI.Delete)
		r.Get("/api/v1/portmaps", portmapAPI.List)
		r.Get("/api/v1/portmaps/new", portmapAPI.New)
		r.Post("/api/v1/portmaps/new", portmapAPI.Create)
		r.Delete("/api/v1/portmaps/:id", portmapAPI.Delete)
		r.Get("/api/v1/routers", routerAPI.List)
		r.Get("/api/v1/routers/new", routerAPI.New)
		r.Post("/api/v1/routers/new", routerAPI.Create)
		r.Delete("/api/v1/routers/:id", routerAPI.Delete)
		r.Get("/api/v1/routers/:id", routerAPI.Edit)
		r.Post("/api/v1/routers/:id", routerAPI.Patch)
		r.Get("/api/v1/secgroups", secgroupAPI.List)
		r.Get("/api/v1/secgroups/new", secgroupAPI.New)
		r.Post("/api/v1/secgroups/new", secgroupAPI.Create)
		r.Delete("/api/v1/secgroups/:id", secgroupAPI.Delete)
		r.Get("/api/v1/secgroups/:id", secgroupAPI.Edit)
		r.Post("/api/v1/secgroups/:id", secgroupAPI.Patch)
		r.Get("/api/v1/secgroups/:sgid/secrules", secruleAPI.List)
		r.Get("/api/v1/secgroups/:sgid/secrules/new", secruleAPI.New)
		r.Post("/api/v1/secgroups/:sgid/secrules/new", secruleAPI.Create)
		r.Delete("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Delete)
		r.Get("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Edit)
		r.Post("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Patch)
	*/
	r.GET("/swagger/api/v1/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("v1")))
	return
}
