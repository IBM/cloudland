package main

import (
	"net/http"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/session"
	_ "github.com/threen134/cdnDashboard/routers"
)

var globalSessions *session.Manager

const TOKENINFO = `tokenInfo`

func init() {
	sessionCfg := &session.ManagerConfig{
		CookieName:      `clandsessionid`,
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		SessionIDPrefix: "cland-session",
	}
	globalSessions, _ = session.NewManager("memory", sessionCfg)
	go globalSessions.GC()
}

func main() {
	AuthFunc := func(ctx *context.Context) {
		if ctx.Request.RequestURI != `/login` {
			token, ok := ctx.Input.Session(TOKENINFO).(string)
			// TODO:  verify token expired and other auth infor
			if !ok {
				ctx.Redirect(http.StatusFound, `/login`)
				logs.Debug("fail to verify token", token)
			}
		}
	}
	beego.BConfig.WebConfig.TemplateLeft = "<<<"
	beego.BConfig.WebConfig.TemplateRight = ">>>"
	beego.BConfig.WebConfig.AutoRender = true
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.SetLogger("file", `{"filename":"logs/test.log"}`)
	beego.InsertFilter("/*", beego.BeforeRouter, AuthFunc)
	beego.Run()
}
