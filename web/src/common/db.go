/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package common

import (
	"context"

	"web/src/dbs"

	"github.com/jinzhu/gorm"
)

var (
	DB = dbs.DB
)

const (
	contextDBKey = "dbs"
)

func GetCtxDB(c context.Context) (context.Context, *gorm.DB) {
	tx := c.Value(contextDBKey)
	if tx != nil {
		return c, tx.(*gorm.DB)
	}
	db := dbs.DB()
	c = context.WithValue(c, contextDBKey, db)
	return c, db
}

func SaveTXtoCtx(c context.Context, db *gorm.DB) context.Context {
	return context.WithValue(c, contextDBKey, db)
}
