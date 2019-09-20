package controllers

import (
	"net/http"

	"github.com/IBM/cloudland/cland-ui/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/session"
)

const TOKENINFO = `tokenInfo`

type MainController struct {
	beego.Controller
	IsLogin bool
}

var globalSessions *session.Manager

// @router  /login [get]
func (c *MainController) Login() {
	if c.IsLogin {
		c.Redirect(`/index`, http.StatusFound)
		return
	}
	c.TplName = "login.html"
	if !c.Ctx.Input.IsPost() {
		return
	}
	flash := beego.NewFlash()
	username := c.GetString("username")
	password := c.GetString("password")
	identity, err := models.Identity()
	if err != nil {
		flash.Error(err.Error())
		flash.Store(&c.Controller)
		return
	}
	c.SetSession("IDENTIY", identity)
	token, err := models.Authenticate(*identity.Versions.Values[0].Links[0].Href, username, password, ``)
	if err != nil {
		flash.Error(err.Error())
		flash.Store(&c.Controller)
		return
	}
	c.IsLogin = true
	logs.Debug(token)
	c.SetSession(TOKENINFO, token)
	c.Redirect(c.URLFor("/index"), 303)
}

// @router  /instances [get]
func (c *MainController) Instances() {
	session := c.StartSession()
	if session.Get(TOKENINFO) == nil {
		c.Redirect(`login.html`, http.StatusFound)
		return
	}
	logs.Debug(session.Get(TOKENINFO))
	c.Layout = "layout.tpl"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["PageContent"] = "instances.tpl"
	c.LayoutSections["SideBar"] = "sidebar.tpl"
	c.LayoutSections["CSS"] = "css.tpl"
	c.LayoutSections["JS"] = "js.tpl"
	c.LayoutSections["FooterContent"] = "footer.tpl"
	c.LayoutSections["TopNavigation"] = "topNavigation.tpl"
	c.LayoutSections["MenuFooter"] = "menuFooter.tpl"
	c.LayoutSections["MenuProfile"] = "menuProfile.tpl"
	c.TplName = "instances.tpl"
}
