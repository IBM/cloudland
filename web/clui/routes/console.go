/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"math/rand"

	"github.com/IBM/cloudland/web/clui/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	"gopkg.in/macaron.v1"
	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
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
	InstanceID string `json:"instanceID"`
	Secret string `json:"secret"`
	jwt.StandardClaims
}

const (
	TokenExpireDuration = time.Hour * 2
)

var SignedSecret = []byte("RedBlue")
//Randomly generate a string of length 10
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func MakeToken(instanceID string, secret string) (string, error) {
	c := TokenClaim{
		InstanceID: instanceID,
		Secret: secret,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "TestIssuer",
		},
	}
	tokenClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := tokenClaim.SignedString(SignedSecret)
	return token, err
}

func ResolveToken(tokenString string) (*TokenClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		return SignedSeret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*TokenClaim); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
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
	tokenString, err := MakeToken(id, RandStringRUnes(10))
	endpoint := viper.GetString("api.endpoint")

	setDB(&Vnc{})
	db := DB()
	//var vncInfo Vnc
	vncInfo := new(Vnc)
	accessAddr := db.First(vncInfo, id).AccessAddress
	accessPort := db.First(vncInfo, id).AccessPort

	consoleURL := fmt.Sprintf("%s/novnc/vnc.html?host=%s&port=%s&autoconnect=true&token=%s", endpoint, accessAddr, accessPort, tokenString)	
	c.Resp.Header().Set("Location", consoleURL)
	c.JSON(301, nil)
	return
}

func (a *ConsoleView) ConsoleResolve(c *macaron.Context, store session.Store) {
	token := c.Params("token")
	tokenClaim, err := ResolveToken(token)
	if err != nil {
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
	}
	log.Println("Get JWT token", token, tokenClaim)

	consoleInfo := &ConsoleInfo{
		Type:      "vnc",
		Address:   "9.115.78.254:5900",
		Insecure:  true,
		TLSTunnel: false,
		Password:  "54321",
	}
	c.JSON(200, consoleInfo)
	return
}
