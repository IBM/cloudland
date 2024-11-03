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
	"reflect"
	"testing"
)

func TestNewOrders(t *testing.T) {
	empty := [][2]string{}
	mapping := [][2]string{
		[2]string{"flavor.name", "flavor"},
		[2]string{"image.id", "image"},
	}
	fixtues := []struct {
		s      string
		orders []string
		m      [][2]string
	}{
		{"+name", []string{"name"}, nil},
		{"-name", []string{"name DESC"}, empty},
		{"-name,created_at", []string{"name DESC", "created_at"}, empty},
		{"-name,created_at", []string{"name DESC", "created_at"}, mapping},
		{"-name,created_at,image.id", []string{"name DESC", "created_at", "image"}, mapping},
		{"-name,created_at,image.id,flavor.name", []string{"name DESC", "created_at", "image", "flavor"}, mapping},
	}
	for _, f := range fixtues {
		var orders []string
		if f.m != nil {
			orders = NewOrders(f.s, f.m...)
		} else {
			orders = NewOrders(f.s)
		}
		if !reflect.DeepEqual(f.orders, orders) {
			t.Fatal(f.orders, orders)
		}
	}
}
