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

func GetContextDB(ctx context.Context) (context.Context, *gorm.DB) {
	tx := ctx.Value(contextDBKey)
	if tx != nil {
		return ctx, tx.(*gorm.DB)
	}
	db := DB()
	ctx = context.WithValue(ctx, contextDBKey, db)
	return ctx, db
}

func SetContextDB(ctx context.Context, db *gorm.DB) context.Context {
	ctx = context.WithValue(ctx, contextDBKey, db)
	return ctx
}

func StartTransaction(ctx context.Context) (context.Context, *gorm.DB, bool) {
	tx := ctx.Value(contextDBKey)
	if tx != nil {
		// returns old transaction
		return ctx, tx.(*gorm.DB), false
	}
	db := DB().Begin()
	ctx = context.WithValue(ctx, contextDBKey, db)
	// returns new transaction
	return ctx, db, true
}

func EndTransaction(ctx context.Context, err error) {
	tx := ctx.Value(contextDBKey)
	if tx != nil {
		db := tx.(*gorm.DB)
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}
}
