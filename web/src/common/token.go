/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"web/src/model"

	"github.com/dgrijalva/jwt-go"
)

type TokenClaim struct {
	OrgID      int64
	Role       model.Role
	InstanceID int    `json:"instanceID"`
	Secret     string `json:"secret"`
	jwt.StandardClaims
}
