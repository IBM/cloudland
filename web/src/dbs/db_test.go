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
	"testing"

	"github.com/jinzhu/gorm"
)

func TestDBMAutomigrate(t *testing.T) {
	type Profile struct {
		gorm.Model
		Name string
	}
	type User struct {
		gorm.Model
		Name      string
		ProfileID uint
		Profile   Profile
	}
	DB().DropTableIfExists("users")
	DB().DropTableIfExists("profiles")

	if err := DB().CreateTable(&Profile{}, &User{}).Error; err != nil {
		t.Fatal(err)
	}
	if !DB().HasTable("users") {
		t.Fatal()
	}
	profile := &Profile{Name: "my profile"}
	user := &User{Name: "me"}
	if err := DB().Create(profile).Error; err != nil {
		t.Fatal(err)
	}
	user.ProfileID = profile.ID
	if err := DB().Create(user).Error; err != nil {
		t.Fatal(err)
	}
	users := []*User{}
	if err := DB().Find(&users).Error; err != nil {
		t.Fatal(err)
	}
	DB().Preload("Profile").Find(&users)
	if len(users) != 1 {
		t.Fatal(len(users))
	}
	if users[0].Profile.ID == 0 {
		t.Fatal()
	}
}

func ExampleTestBool() {
	type MyTestBool struct {
		gorm.Model
		Slims bool
	}
	db := DB()
	db.AutoMigrate(&MyTestBool{})
	db.Unscoped().Delete(&MyTestBool{})
	db.Create(&MyTestBool{
		Slims: false,
	})
	db.Create(&MyTestBool{})
	db.Create(&MyTestBool{Slims: true})
	count := 0
	db.Model(&MyTestBool{}).Where(map[string]interface{}{"slims": false}).Count(&count)
	fmt.Println(count)
	// Output:
	// 2
}

func ExampleTestHasOne() {
	type TestAddress struct {
		gorm.Model
		Address1 string
	}
	type TestEmail struct {
		gorm.Model
		Email string
	}
	type TestLanguage struct {
		gorm.Model
		Name string
	}
	type TestUser struct {
		gorm.Model
		Name             string
		Age              int32 `gorm:"default:-1"`
		BillingAddressID uint
		BillingAddress   TestAddress
		Emails           []TestEmail
		Languages        []TestLanguage
	}
	user := &TestUser{
		Name:           "jinzhu",
		BillingAddress: TestAddress{Address1: "Billing Address - Address 1"},
		Emails: []TestEmail{
			{Email: "jinzhu@example.com"},
			{Email: "jinzhu-2@example@example.com"},
		},
		Languages: []TestLanguage{
			{Name: "ZH"},
			{Name: "EN"},
		},
	}
	db := DB()
	db.Where(user).First(user)
	fmt.Println(user.Age)
	db.AutoMigrate(&TestAddress{}, &TestEmail{}, &TestLanguage{}, &TestUser{})
	db.Create(user)
	out := &TestUser{
		BillingAddressID: 1,
	}
	fmt.Println(db.First(out, out).Error)
	user2 := &TestUser{
		Name:             "me",
		BillingAddressID: 1,
		BillingAddress: TestAddress{
			Model:    gorm.Model{ID: 1},
			Address1: "Billing Address - Address 2",
		},
	}
	db.Create(user2)
	db.Model(&TestUser{}).Update("billing_address_id", 0)
	fmt.Println(user.BillingAddressID)
	db.Delete(user)
	billingAddress := &TestAddress{}
	db.First(billingAddress, 1)
	fmt.Println(billingAddress.Address1)
	// -1
	// Output: <nil>
	// 1
	// Billing Address - Address 2
}

func TestDialect(t *testing.T) {
	db := DB()
	dialect := db.Dialect().GetName()
	if dialect == "" {
		t.Fatal()
	}
}
