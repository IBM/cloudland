/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package routes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	. "web/src/common"
	"web/src/model"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-macaron/session"
	"github.com/spf13/viper"
	"golang.org/x/crypto/sha3"
	"gopkg.in/macaron.v1"
)

var (
	consoleView = &ConsoleView{}
)

type ConsoleView struct{}

const (
	TokenExpireDuration = time.Hour * 2
)

// Randomly generate a string of length 10
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

func MakeToken(ctx context.Context, instance *model.Instance) (token string, err error) {
	memberShip := GetMemberShip(ctx)
	permit := memberShip.CheckPermission(model.Writer)
	if !permit {
		log.Println("Not authorized to create interface in public subnet")
		err = fmt.Errorf("Not authorized")
		return
	}
	secret := RandomStr()
	tkClaim := TokenClaim{
		OrgID:      memberShip.OrgID,
		Role:       memberShip.Role,
		InstanceID: int(instance.ID),
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
		Instance:   instance.ID,
		Type:       "vnc",
		HashSecret: hashSecret,
	}
	err = db.Where("instance = ?", instance.ID).Assign(console).FirstOrCreate(&model.Console{}).Error
	if err != nil {
		log.Println("Failed to make console record ", err)
		return
	}
	tokenClaim := jwt.NewWithClaims(jwt.SigningMethodHS256, tkClaim)
	token, err = tokenClaim.SignedString(SignedSeret)
	return
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
	ctx := c.Req.Context()
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
	instance, err := instanceAdmin.Get(ctx, int64(instanceID))
	if err != nil {
		code := http.StatusBadRequest
		c.Error(code, http.StatusText(code))
		return
	}
	tokenString, err := MakeToken(ctx, instance)
	if err != nil {
		log.Println("failed to make token", err)
		code := http.StatusInternalServerError
		c.Error(code, http.StatusText(code))
		return
	}
	accessAddr := viper.GetString("console.host")
	accessPort := viper.GetInt("console.port")
	consoleURL := fmt.Sprintf("/novnc/vnc.html?host=%s&port=%d&autoconnect=true&encrypt=true&path=websockify?token=%s", accessAddr, accessPort, tokenString)
	c.Resp.Header().Set("Location", consoleURL)
	c.JSON(301, nil)
	return
}
