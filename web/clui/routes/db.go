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
	"context"

	"github.com/IBM/cloudland/web/sca/dbs"
	"github.com/jinzhu/gorm"
)

var (
	DB = dbs.DB
)

const (
	contextDBKey = "dbs"
)

func getCtxDB(c context.Context) (context.Context, *gorm.DB) {
	tx := c.Value(contextDBKey)
	if tx != nil {
		return c, tx.(*gorm.DB)
	}
	db := DB()
	c = context.WithValue(c, contextDBKey, db)
	return c, db
}

func saveTXtoCtx(c context.Context, db *gorm.DB) context.Context {
	return context.WithValue(c, contextDBKey, db)
}
