/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"

	_ "web/docs"

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
		authGroup.GET("/api/v1/hypers", hyperAPI.List)
		authGroup.GET("/api/v1/hypers/:name", hyperAPI.Get)

		authGroup.GET("/api/v1/users", userAPI.List)
		authGroup.POST("/api/v1/users", userAPI.Create)
		authGroup.GET("/api/v1/users/:id", userAPI.Get)
		authGroup.DELETE("/api/v1/users/:id", userAPI.Delete)
		authGroup.PATCH("/api/v1/users/:id", userAPI.Patch)

		authGroup.GET("/api/v1/orgs", orgAPI.List)
		authGroup.POST("/api/v1/orgs", orgAPI.Create)
		authGroup.GET("/api/v1/orgs/:id", orgAPI.Get)
		authGroup.DELETE("/api/v1/orgs/:id", orgAPI.Delete)
		authGroup.PATCH("/api/v1/orgs/:id", orgAPI.Patch)

		authGroup.GET("/api/v1/vpcs", vpcAPI.List)
		authGroup.POST("/api/v1/vpcs", vpcAPI.Create)
		authGroup.GET("/api/v1/vpcs/:id", vpcAPI.Get)
		authGroup.DELETE("/api/v1/vpcs/:id", vpcAPI.Delete)
		authGroup.PATCH("/api/v1/vpcs/:id", vpcAPI.Patch)

		authGroup.GET("/api/v1/subnets", subnetAPI.List)
		authGroup.POST("/api/v1/subnets", subnetAPI.Create)
		authGroup.GET("/api/v1/subnets/:id", subnetAPI.Get)
		authGroup.DELETE("/api/v1/subnets/:id", subnetAPI.Delete)
		authGroup.PATCH("/api/v1/subnets/:id", subnetAPI.Patch)

		authGroup.GET("/api/v1/security_groups", secgroupAPI.List)
		authGroup.POST("/api/v1/security_groups", secgroupAPI.Create)
		authGroup.GET("/api/v1/security_groups/:id", secgroupAPI.Get)
		authGroup.DELETE("/api/v1/security_groups/:id", secgroupAPI.Delete)
		authGroup.PATCH("/api/v1/security_groups/:id", secgroupAPI.Patch)

		authGroup.GET("/api/v1/security_groups/:id/rules", secruleAPI.List)
		authGroup.POST("/api/v1/security_groups/:id/rules", secruleAPI.Create)
		authGroup.GET("/api/v1/security_groups/:id/rules/:rule_id", secruleAPI.Get)
		authGroup.DELETE("/api/v1/security_groups/:id/rules/:rule_id", secruleAPI.Delete)
		authGroup.PATCH("/api/v1/security_groups/:id/rules/:rule_id", secruleAPI.Patch)

		authGroup.GET("/api/v1/floating_ips", floatingIpAPI.List)
		authGroup.POST("/api/v1/floating_ips", floatingIpAPI.Create)
		authGroup.GET("/api/v1/floating_ips/:id", floatingIpAPI.Get)
		authGroup.DELETE("/api/v1/floating_ips/:id", floatingIpAPI.Delete)
		authGroup.PATCH("/api/v1/floating_ips/:id", floatingIpAPI.Patch)

		authGroup.GET("/api/v1/keys", keyAPI.List)
		authGroup.POST("/api/v1/keys", keyAPI.Create)
		authGroup.GET("/api/v1/keys/:id", keyAPI.Get)
		authGroup.DELETE("/api/v1/keys/:id", keyAPI.Delete)
		authGroup.PATCH("/api/v1/keys/:id", keyAPI.Patch)

		authGroup.GET("/api/v1/flavors", flavorAPI.List)
		authGroup.POST("/api/v1/flavors", flavorAPI.Create)
		authGroup.GET("/api/v1/flavors/:id", flavorAPI.Get)
		authGroup.DELETE("/api/v1/flavors/:id", flavorAPI.Delete)
		authGroup.PATCH("/api/v1/flavors/:id", flavorAPI.Patch)

		authGroup.GET("/api/v1/images", imageAPI.List)
		authGroup.POST("/api/v1/images", imageAPI.Create)
		authGroup.GET("/api/v1/images/:id", imageAPI.Get)
		authGroup.DELETE("/api/v1/images/:id", imageAPI.Delete)
		authGroup.PATCH("/api/v1/images/:id", imageAPI.Patch)

		authGroup.GET("/api/v1/volumes", volumeAPI.List)
		authGroup.POST("/api/v1/volumes", volumeAPI.Create)
		authGroup.GET("/api/v1/volumes/:id", volumeAPI.Get)
		authGroup.DELETE("/api/v1/volumes/:id", volumeAPI.Delete)
		authGroup.PATCH("/api/v1/volumes/:id", volumeAPI.Patch)

		authGroup.GET("/api/v1/instances", instanceAPI.List)
		authGroup.POST("/api/v1/instances", instanceAPI.Create)
		authGroup.GET("/api/v1/instances/:id", instanceAPI.Get)
		authGroup.DELETE("/api/v1/instances/:id", instanceAPI.Delete)
		authGroup.PATCH("/api/v1/instances/:id", instanceAPI.Patch)

		authGroup.GET("/api/v1/instances/:id/interfaces", interfaceAPI.List)
		authGroup.POST("/api/v1/instances/:id/interfaces", interfaceAPI.Create)
		authGroup.GET("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Get)
		authGroup.DELETE("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Delete)
		authGroup.PATCH("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Patch)
		/*
			authGroup.Post("/api/v1/instances/:id/console", consoleAPI.ConsoleURL)
			authGroup.Get("/api/v1/consoleresolver/token/:token", consoleAPI.ConsoleResolve)
		*/
	}
	r.GET("/swagger/api/v1/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("v1")))
	return
}
