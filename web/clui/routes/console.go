/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/cloudland/web/clui/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
)

var (
	consoleAdmin = &ConsoleAdmin{}
	consoleView  = &ConsoleView{}
)

type ConsoleAdmin struct{}
type ConsoleView struct{}

type ConsoleInfo struct {
	Type      string `json:"type"`
	Address   string `json:"address"`
	Insecure  bool   `json:"insecure"`
	TLSTunnel bool   `json:"tlsTunnel"`
	Password  string `json:"password"`
}
type TokenClaim struct {
	InstanceID int    `json:"instanceID"`
	Secret     string `json:"secret"`
	jwt.StandardClaims
}

const (
	TokenExpireDuration = time.Hour * 2
)

var SignedSeret = []byte("Red B")

//Randomly generate a string of length 10
func RandomStr() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func MakeToken(instanceID int, secret string) (string, error) {
	c := TokenClaim{
		InstanceID: instanceID,
		Secret:     secret,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "TestIssuer",
		},
	}
	tokenClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := tokenClaim.SignedString(SignedSeret)
	return token, err
}

func ResolveToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		return SignedSeret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*TokenClaim); ok && token.Valid {
		return claims.InstanceID, nil
	}
	return 0, errors.New("invalid token")
}

func (a *ConsoleView) ConsoleURL(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	id := c.Params("id")
	if id == "" {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	instanceID, err := strconv.Atoi(id)
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	permit, err := memberShip.CheckOwner(model.Reader, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	tokenString, err := MakeToken(instanceID, RandomStr())
	if err != nil {
		log.Println("failed to make token", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	endpoint := viper.GetString("api.endpoint")
	accessAddr := viper.GetString("console.host")
	accessPort := viper.GetInt("console.port")
	consoleURL := fmt.Sprintf("%s/novnc/vnc.html?host=%s&port=%d&autoconnect=true&token=%s", endpoint, accessAddr, accessPort, tokenString)
	c.Resp.Header().Set("Location", consoleURL)
	c.JSON(301, nil)
	return
}

func (a *ConsoleView) ConsoleResolve(c *macaron.Context, store session.Store) {
	memberShip := GetMemberShip(c.Req.Context())
	token := c.Params("token")
	log.Println("Get JWT token", token)
	instanceID, err := ResolveToken(token)
	if err != nil {
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
	}
	permit, err := memberShip.CheckOwner(model.Writer, "instances", int64(instanceID))
	if !permit {
		log.Println("Not authorized for this operation")
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
	}
	db := DB()
	vnc := &model.Vnc{InstanceID: int64(instanceID)}
	err = db.Where(vnc).Take(vnc).Error
	if err != nil {
		log.Println("VNC query failed", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}

	accessPass := vnc.Passwd
	address := fmt.Sprintf("%s:%s", vnc.AccessAddress, vnc.AccessPort)
	consoleInfo := &ConsoleInfo{
		Type:      "vnc",
		Address:   address,
		Insecure:  true,
		TLSTunnel: false,
		Password:  accessPass,
	}

	c.JSON(200, consoleInfo)
	return
}
