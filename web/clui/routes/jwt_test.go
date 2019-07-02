/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package routes

import (
	"crypto/rsa"
	"testing"

	"github.com/IBM/cloudland/web/clui/model"
	jwt "github.com/dgrijalva/jwt-go"
)

func TestSigningMethodRS256(t *testing.T) {
	idrsa := testdata["key.private"].(*rsa.PrivateKey)
	idrsaPub := testdata["key.public"].(*rsa.PublicKey)
	claim, _, _ := NewClaims("admin", "admin", 1, 1, model.Reader)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claim)
	s, err := token.SignedString(idrsa)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
	claims := &HypercubeClaims{}
	token, err = jwt.ParseWithClaims(s, claims, func(in *jwt.Token) (key interface{}, err error) {
		key = idrsaPub
		return
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := token.Claims.Valid(); err != nil {
		t.Fatal(token)
	}
	t.Log(token)
}

func BenchmarkParseWithClaims(b *testing.B) {
	publicKey := testdata["key.public"].(*rsa.PublicKey)
	s := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhZG1pbiIsImV4cCI6MTU0NTgxNDQ5OCwianRpIjoiNDM0MjQxNTc3NjM2MTY4NDU3MiIsImlhdCI6MTU0NTgwNzI5OCwiaXNzIjoiaHlwZXJjdWJlIiwibmJmIjoxNTQ1ODA3Mjk4LCJzdWIiOiJhZG1pbiIsInVpZCI6MSwib2lkIjoxfQ.leHLR-wvQTyua0RtIpy1MxSoeJFI3aqtF8jKuqp6Zs4OisIa8wfnA2H1lUfhn9sFn686i5gh7J1r02uiAuXkJd3TWGTVXOwveAc03CnU1p8zybXsSITNKblSLldf8fY6Akhi6HwvVCO8FYLQ_4LKBMg5bweQWdMD8OxC7YvAOznMOf-PQ0ipQfL_Ri4O8Olqs9xDE5O7Lm4oENKB3iikme0QFN6zgzduzxKLd-8whv6eeu54Ig2J-fX8atjhCQWDqExlOTLfpQgtGhgaitzjjzIajuAiekR-nV4suzUo8gctJXv-YbN2A4UWm_jyouQ2ZrisxjGST9_g8GiF0rbLyA`
	for i := 0; i < b.N; i++ {
		claims := &HypercubeClaims{}
		_, err := jwt.ParseWithClaims(s, claims, func(in *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			b.Fatal(err)
		}
		if claims.OID != 1 || claims.UID != 1 {
			b.Fatal(claims)
		}
	}
}
