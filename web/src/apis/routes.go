/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	_ "web/docs"
	"web/src/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var logger = log.MustGetLogger("apis")

func Run() (err error) {
	logger.Info("Start to run cloudland api service")
	r := Register()
	cert := viper.GetString("rest.cert")
	key := viper.GetString("rest.key")
	listen := viper.GetString("rest.listen")
	logger.Infof("cert: %s, key: %s\n", cert, key)
	if cert != "" && key != "" {
		logger.Infof("Running https service isten on %s\n", listen)
		r.RunTLS(listen, cert, key)
	} else {
		logger.Infof("Running http service on %s\n", listen)
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
	r = gin.Default()

	r.POST("/api/v1/login", userAPI.LoginPost)
	r.GET("/api/v1/version", versionAPI.Get)
	authGroup := r.Group("").Use(Authorize())
	{
		//authGroup.GET("/api/v1/version", versionAPI.Get)
		authGroup.GET("/api/v1/hypers", hyperAPI.List)
		authGroup.GET("/api/v1/hypers/:name", hyperAPI.Get)

		authGroup.GET("/api/v1/migrations", migrationAPI.List)
		authGroup.POST("/api/v1/migrations", migrationAPI.Create)
		authGroup.GET("/api/v1/migrations/:id", migrationAPI.Get)

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
		authGroup.GET("/api/v1/flavors/:name", flavorAPI.Get)
		authGroup.DELETE("/api/v1/flavors/:name", flavorAPI.Delete)

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

		authGroup.POST("/api/v1/instances/:id/set_user_password", instanceAPI.SetUserPassword)

		authGroup.POST("/api/v1/instances/:id/console", consoleAPI.Create)

		authGroup.POST("/api/v1/instances/:id/reinstall", instanceAPI.Reinstall)

		authGroup.GET("/api/v1/instances/:id/interfaces", interfaceAPI.List)
		authGroup.POST("/api/v1/instances/:id/interfaces", interfaceAPI.Create)
		authGroup.GET("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Get)
		authGroup.DELETE("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Delete)
		authGroup.PATCH("/api/v1/instances/:id/interfaces/:interface_id", interfaceAPI.Patch)

		metricsGroup := authGroup.(*gin.RouterGroup).Group("/api/v1/metrics")
		{
			metricsGroup.POST("/instances/cpu/his_data", monitorAPI.GetCPU)
			metricsGroup.POST("/instances/disk/his_data", monitorAPI.GetDisk)
			metricsGroup.POST("/instances/memory/his_data", monitorAPI.GetMemory)
			metricsGroup.POST("/instances/network/his_data", monitorAPI.GetNetwork)
			metricsGroup.POST("/instances/traffic/his_data", monitorAPI.GetTraffic)
			metricsGroup.POST("/instances/volume/his_data", monitorAPI.GetVolume)
		}

	}

	r.GET("/swagger/api/v1/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("v1")))
	return
}
