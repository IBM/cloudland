/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package rpcs

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	. "web/src/common"
	"web/src/model"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/sha3"
	"gopkg.in/macaron.v1"
)

var (
	consoleAdmin = &ConsoleAdmin{}
)

type ConsoleAdmin struct{}

type ConsoleInfo struct {
	Type      string `json:"type"`
	Address   string `json:"address"`
	Insecure  bool   `json:"insecure"`
	TLSTunnel bool   `json:"tlsTunnel"`
	Password  string `json:"password"`
}

func ResolveToken(tokenString string) (int, *MemberShip, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		return SignedSeret, nil
	})
	if err != nil || token == nil {
		return 0, nil, err
	}
	claims, ok := token.Claims.(*TokenClaim)
	if !ok || !token.Valid {
		return 0, nil, errors.New("invalid token")
	}
	instanceID := claims.InstanceID
	console := &model.Console{Instance: int64(instanceID)}
	err = DB().Where(console).Take(console).Error
	if err != nil {
		return 0, nil, err
	}
	tokenHash := make([]byte, 32)
	data := sha3.NewShake256()
	data.Write([]byte(claims.Secret))
	data.Read(tokenHash)
	hashSecret := fmt.Sprintf("%x", tokenHash)
	if hashSecret != console.HashSecret {
		return 0, nil, errors.New("Secret can not pass validation")
	}
	memberShip := &MemberShip{
		OrgID: claims.OrgID,
		Role:  claims.Role,
	}
	return instanceID, memberShip, nil
}

func (a *ConsoleAdmin) ConsoleResolve(c *macaron.Context) {
	token := c.Params("token")
	logger.Debug("Get JWT token", token)
	instanceID, memberShip, err := ResolveToken(token)
	if err != nil {
		logger.Error("Unable to resolve token", err)
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		logger.Error("Not authorized for this operation")
		err = fmt.Errorf("Not authorized")
		return
	}
	db := DB()
	instance := &model.Instance{Model: model.Model{ID: int64(instanceID)}}
	err = db.Take(instance).Error
	if err != nil {
		logger.Error("Failed to get instance", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}

	accessPass, err := password.Generate(8, 2, 0, false, false)
	if err != nil {
		logger.Error("Failed to generate password")
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	vnc := &model.Vnc{InstanceID: int64(instanceID)}
	err = db.Where(vnc).Delete(vnc).Error
	if err != nil {
		logger.Error("VNC record deletion failed", err)
	}
	control := fmt.Sprintf("inter=%d", instance.Hyper)
	command := fmt.Sprintf("/opt/cloudland/scripts/backend/set_vnc_passwd.sh '%d' '%s'", instance.ID, accessPass)
	err = HyperExecute(c.Req.Context(), control, command)
	if err != nil {
		logger.Error("Set vnc password execution failed", err)
		return
	}

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		err = db.Where(vnc).Take(vnc).Error
		if err == nil {
			logger.Error("get VNC record successfully, i = ", i)
			break
		}
	}
	if vnc.LocalAddress == "" {
		logger.Error("get VNC record successfully", err)
		c.JSON(http.StatusInternalServerError, &APIError{ErrorMessage: "Internal error"})
	}
	address := fmt.Sprintf("%s:%d", vnc.LocalAddress, vnc.LocalPort)
	consoleInfo := &ConsoleInfo{
		Type:      "vnc",
		Address:   address,
		Insecure:  true,
		TLSTunnel: false,
		Password:  accessPass,
	}

	c.JSON(http.StatusOK, consoleInfo)
	return
}
