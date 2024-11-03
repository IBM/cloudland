/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"

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
	r.Use(gin.Recovery())

	r.POST("/api/v1/login", userAPI.LoginPost)
	authGroup := r.Group("").Use(Authorize())
	{
		/*
			authGroup.Get("/api/v1/hypers", hyperAPI.List)
			authGroup.Get("/api/v1/users", userAPI.List)
			authGroup.Get("/api/v1/users/:id", userAPI.Edit)
			authGroup.Post("/api/v1/users/:id", userAPI.Patch)
			authGroup.Get("/api/v1/users/:id/chorg", userAPI.Change)
			authGroup.Delete("/api/v1/users/:id", userAPI.Delete)
			authGroup.Get("/api/v1/users/new", userAPI.New)
			authGroup.Post("/api/v1/users/new", userAPI.Create)
			authGroup.Get("/api/v1/orgs", orgAPI.List)
			authGroup.Get("/api/v1/orgs/:id", orgAPI.Edit)
			authGroup.Post("/api/v1/orgs/:id", orgAPI.Patch)
			authGroup.Delete("/api/v1/orgs/:id", orgAPI.Delete)
			authGroup.Get("/api/v1/orgs/new", orgAPI.New)
			authGroup.Post("/api/v1/orgs/new", orgAPI.Create)
		*/
		authGroup.GET("/api/v1/instances", instanceAPI.List)
		/*
			authGroup.Post("/api/v1/instances", instanceAPI.Create)
			authGroup.Get("/api/v1/instances/:id", instanceAPI.Get)
			authGroup.Delete("/api/v1/instances/:id", instanceAPI.Delete)
			authGroup.Get("/api/v1/instances/:id", instanceAPI.Edit)
			authGroup.Post("/api/v1/instances/:id", instanceAPI.Patch)
			authGroup.Post("/api/v1/instances/:id/console", consoleAPI.ConsoleURL)
			authGroup.Get("/api/v1/consoleresolver/token/:token", consoleAPI.ConsoleResolve)
			authGroup.Get("/api/v1/interfaces/:id", interfaceAPI.Edit)
			authGroup.Post("/api/v1/interfaces/:id", interfaceAPI.Patch)
			authGroup.Post("/api/v1/interfaces/new", interfaceAPI.Create)
			authGroup.Delete("/api/v1/interfaces/:id", interfaceAPI.Delete)
			authGroup.Get("/api/v1/flavors", flavorAPI.List)
			authGroup.Get("/api/v1/flavors/new", flavorAPI.New)
			authGroup.Post("/api/v1/flavors/new", flavorAPI.Create)
			authGroup.Delete("/api/v1/flavors/:id", flavorAPI.Delete)
			authGroup.Get("/api/v1/images", imageAPI.List)
			authGroup.Get("/api/v1/images/new", imageAPI.New)
			authGroup.Post("/api/v1/images/new", imageAPI.Create)
			authGroup.Delete("/api/v1/images/:id", imageAPI.Delete)
			authGroup.Get("/api/v1/volumes", volumeAPI.List)
			authGroup.Get("/api/v1/volumes/new", volumeAPI.New)
			authGroup.Post("/api/v1/volumes/new", volumeAPI.Create)
			authGroup.Delete("/api/v1/volumes/:id", volumeAPI.Delete)
			authGroup.Get("/api/v1/volumes/:id", volumeAPI.Edit)
			authGroup.Post("/api/v1/volumes/:id", volumeAPI.Patch)
			authGroup.Get("/api/v1/subnets", subnetAPI.List)
			authGroup.Get("/api/v1/subnets/new", subnetAPI.New)
			authGroup.Post("/api/v1/subnets/new", subnetAPI.Create)
			authGroup.Delete("/api/v1/subnets/:id", subnetAPI.Delete)
			authGroup.Get("/api/v1/subnets/:id", subnetAPI.Edit)
			authGroup.Post("/api/v1/subnets/:id", subnetAPI.Patch)
			authGroup.Get("/api/v1/keys", keyAPI.List)
			authGroup.Get("/api/v1/keys/new", keyAPI.New)
			authGroup.Post("/api/v1/keys/new", keyAPI.Create)
			authGroup.Post("/api/v1/keys/confirm", keyAPI.Confirm)
			authGroup.Delete("/api/v1/keys/:id", keyAPI.Delete)
			authGroup.Get("/api/v1/floatingips", floatingipAPI.List)
			authGroup.Get("/api/v1/floatingips/new", floatingipAPI.New)
			authGroup.Post("/api/v1/floatingips/new", floatingipAPI.Create)
			authGroup.Post("/api/v1/floatingips/assign", floatingipAPI.Assign)
			authGroup.Delete("/api/v1/floatingips/:id", floatingipAPI.Delete)
			authGroup.Get("/api/v1/portmaps", portmapAPI.List)
			authGroup.Get("/api/v1/portmaps/new", portmapAPI.New)
			authGroup.Post("/api/v1/portmaps/new", portmapAPI.Create)
			authGroup.Delete("/api/v1/portmaps/:id", portmapAPI.Delete)
			authGroup.Get("/api/v1/routers", routerAPI.List)
			authGroup.Get("/api/v1/routers/new", routerAPI.New)
			authGroup.Post("/api/v1/routers/new", routerAPI.Create)
			authGroup.Delete("/api/v1/routers/:id", routerAPI.Delete)
			authGroup.Get("/api/v1/routers/:id", routerAPI.Edit)
			authGroup.Post("/api/v1/routers/:id", routerAPI.Patch)
			authGroup.Get("/api/v1/secgroups", secgroupAPI.List)
			authGroup.Get("/api/v1/secgroups/new", secgroupAPI.New)
			authGroup.Post("/api/v1/secgroups/new", secgroupAPI.Create)
			authGroup.Delete("/api/v1/secgroups/:id", secgroupAPI.Delete)
			authGroup.Get("/api/v1/secgroups/:id", secgroupAPI.Edit)
			authGroup.Post("/api/v1/secgroups/:id", secgroupAPI.Patch)
			authGroup.Get("/api/v1/secgroups/:sgid/secrules", secruleAPI.List)
			authGroup.Get("/api/v1/secgroups/:sgid/secrules/new", secruleAPI.New)
			authGroup.Post("/api/v1/secgroups/:sgid/secrules/new", secruleAPI.Create)
			authGroup.Delete("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Delete)
			authGroup.Get("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Edit)
			authGroup.Post("/api/v1/secgroups/:sgid/secrules/:id", secruleAPI.Patch)
		*/
	}
	r.GET("/swagger/api/v1/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("v1")))
	return
}
