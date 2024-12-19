/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package apis

import (
	"log"
	"net/http"

	. "web/src/common"
	"web/src/routes"

	"github.com/gin-gonic/gin"
)

const (
	TokenType = "bearer"
	AppName   = "Cloudland"
)

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("Authorization")
		if tokenStr == "" {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid Token", nil)
			c.Abort()
			return
		}
		tokenStr = tokenStr[len(TokenType)+1:]
		_, claims, err := routes.ParseToken(tokenStr)
		if err != nil {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid Token", err)
			c.Abort()
			return
		}
		if claims.Issuer != AppName {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid Token", nil)
			c.Abort()
			return
		}

		reqUser := claims.Audience
		reqOrg := claims.Subject
		realUser := c.Request.Header.Get("X-Resource-User")
		realOrg := c.Request.Header.Get("X-Resource-Org")
		if realUser != "" || realOrg != "" {
			if reqUser != "admin" {
				ErrorResponse(c, http.StatusUnauthorized, "Not authorized to change resource owner", nil)
				c.Abort()
				return
			}
		}
		if realUser == "" {
			realUser = reqUser
			realOrg = reqOrg
		}
		user, err := userAdmin.GetUserByName(realUser)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid resource user", err)
			c.Abort()
			return
		}
		if realOrg == "" {
			realOrg = realUser
		}
		org, err := orgAdmin.GetOrgByName(realOrg)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid resource org", err)
			c.Abort()
			return
		}
		memberShip, err := GetDBMemberShip(user.ID, org.ID)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid resource user with org membership", err)
			c.Abort()
			return
		}
		/*
			if realUser == "admin" {
				memberShip.Role = model.Admin
			}
		*/
		log.Printf("MemberShip: %v\n", memberShip)
		ctx := memberShip.SetContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
