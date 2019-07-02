/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	"testing"

	"github.com/jinzhu/gorm"
)

func TestQuery(t *testing.T) {
	type TestQuery01 struct {
		gorm.Model
		Name string
		Age  int32
	}
	db := DB()
	db.AutoMigrate(&TestQuery01{})
	db.Create(&TestQuery01{
		Name: "Test01",
	})
	db.Create(&TestQuery01{
		Name: "Test02",
	})
	rs, err := query("select * from test_query01")
	if err != nil || len(rs) != 3 {
		t.Fatal(rs, err)
	}
	affected, err := execSql("update test_query01 set age = 10")
	if err != nil || affected != 2 {
		t.Fatal(affected, err)
	}
	affected, err = execSql("delete from test_query01 where age = 10")
	if err != nil || affected != 2 {
		t.Fatal(affected, err)
	}
	rs, err = query("select * from test_query01")
	if err != nil || len(rs) != 1 {
		t.Fatal(rs, err)
	}
}
