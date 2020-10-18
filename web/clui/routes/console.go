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
	"golang.org/x/crypto/sha3"
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
	OrgID      int64
	Role       model.Role
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

func MakeToken(instanceID int, secret string, memberShip *MemberShip) (string, error) {
	tkClaim := TokenClaim{
		OrgID:      memberShip.OrgID,
		Role:       memberShip.Role,
		InstanceID: instanceID,
		Secret:     secret,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
		},
	}
	tokenHash := make([]byte, 32)
	data := sha3.NewShake256()
	data.Write([]byte(secret))
	data.Read(tokenHash)
	hashSecret := fmt.Sprintf("%x", tokenHash)
	db := DB()
	console := &model.Console{
		Instance:   int64(instanceID),
		Type:       "vnc",
		HashSecret: hashSecret,
	}
	err := db.Where("instance = ?", instanceID).Assign(console).FirstOrCreate(&model.Console{}).Error
	if err != nil {
		return "", err
	}
	tokenClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, tkClaim)
	token, err := tokenClaim.SignedString(SignedSeret)
	if err != nil {
		return "", err
	}
	return token, nil
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
	tokenString, err := MakeToken(instanceID, RandomStr(), memberShip)
	if err != nil {
		log.Println("failed to make token", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	endpoint := viper.GetString("api.endpoint")
	accessAddr := viper.GetString("console.host")
	accessPort := viper.GetInt("console.port")
	consoleURL := fmt.Sprintf("%s/novnc/vnc.html?host=%s&port=%d&autoconnect=true&encrypt=true&path=websockify?token=%s", endpoint, accessAddr, accessPort, tokenString)
	c.Resp.Header().Set("Location", consoleURL)
	c.JSON(301, nil)
	return
}

func (a *ConsoleView) ConsoleResolve(c *macaron.Context, store session.Store) {
	token := c.Params("token")
	log.Println("Get JWT token", token)
	instanceID, memberShip, err := ResolveToken(token)
	if err != nil {
		log.Println("Unable to resolve token", err)
		code := http.StatusUnauthorized
		c.Error(code, http.StatusText(code))
		return
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
	address := fmt.Sprintf("%s:%d", vnc.LocalAddress, vnc.LocalPort)
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
