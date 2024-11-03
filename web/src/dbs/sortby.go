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
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

// Sortby Sort by `s` defined at api handbook collections sorting
func Sortby(db *gorm.DB, s string, m ...[2]string) *gorm.DB {
	orders := NewOrders(s, m...)
	for _, order := range orders {
		db = db.Order(order)
	}
	return db
}

// NewOrders new orders
// s: sort string defined in api handbook collections sorting
// m: mappings
func NewOrders(s string, m ...[2]string) (orders []string) {
	if s == "" {
		return
	}
	mapping := func(k string) (v string) {
		for _, p := range m {
			if k == p[0] {
				v = p[1]
				return
			}
		}
		return k
	}
	items := strings.Split(s, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = mapping(item)
		if item == "" {
			continue
		}
		c := item[0]
		switch c {
		case '-':
			item = fmt.Sprintf("%s DESC", item[1:])
		case '+':
			item = item[1:]
		}
		orders = append(orders, item)
	}
	return
}
