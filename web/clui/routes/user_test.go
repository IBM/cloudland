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
	"strconv"
	"testing"

	"github.com/IBM/cloudland/web/clui/model"
)

func TestUserAdminCreate(t *testing.T) {
	userAdmin.Delete(0)
	defer userAdmin.Delete(0) // delete all
	username := "admin"
	password := "admin"
	user, err := userAdmin.Create(username, password)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal(user)
	}
	if user.Password == password {
		t.Fatal(user)
	}
}

func TestUserAdminValidate(t *testing.T) {
	userAdmin.Delete(0)
	defer userAdmin.Delete(0) // delete all
	username := "admin"
	password := "admin"
	user, err := userAdmin.Create(username, password)
	if err != nil {
		t.Fatal(err)
	}
	userID := user.ID
	if userID == 0 {
		t.Fatal(user)
	}
	user, err = userAdmin.Validate(username, password)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != userID {
		t.Fatal(user)
	}

}

func TestUserAdminAccessToken(t *testing.T) {
	userAdmin.Delete(0)
	orgAdmin.Delete(0)
	defer userAdmin.Delete(0)                       // delete all users
	defer orgAdmin.Delete(0)                        // delete all orgs
	user, err := userAdmin.Create("admin", "admin") // create admin user
	if err != nil {
		t.Fatal(err)
	}

	_, err = orgAdmin.Create("admin", strconv.FormatInt(user.ID, 10))
	if err != nil {
		t.Fatal(err)
	}
	oid, role, accessToken, _, _, err := userAdmin.AccessToken(user.ID, "admin", "admin")
	if err != nil || oid == 0 || role == model.None {
		t.Fatal(err, oid, role)
	}
	t.Log(accessToken)
}
