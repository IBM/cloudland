package routers

import (
	"github.com/astaxie/beego"
	"github.com/threen134/cdnDashboard/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{}, "get:Instances")
	beego.Router("/index", &controllers.MainController{}, "get:Instances")
	beego.Router("/login", &controllers.MainController{}, "get,post:Login")
	beego.Router("/instances", &controllers.MainController{}, "get:Instances")
}
