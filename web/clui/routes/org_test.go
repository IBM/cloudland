package routes

import (
	"fmt"
	"testing"

	"github.com/IBM/cloudland/web/clui/model"
	"github.com/IBM/cloudland/web/sca/dbs"
)

func TestRole(t *testing.T) {
	member := &model.Member{}
	role := fmt.Sprint(member.Role)
	if role != "None" {
		t.Fatal(role)
	}
	member.Role = model.Owner
	role = fmt.Sprint(member.Role)
	if role != "Owner" {
		t.Fatal(role)
	}
}

func TestOrgCreate(t *testing.T) {
	userAdmin.Delete(0)
	orgAdmin.Delete(0)
	username := "admin"
	password := "admin"
	admin, err := userAdmin.Create(username, password)
	if err != nil {
		t.Fatal(err)
	}
	owner := admin.ID
	defer userAdmin.Delete(0)
	defer orgAdmin.Delete(0)
	org, err := orgAdmin.Create("admin", owner)
	if err != nil {
		t.Fatal(err)
	}
	orgID := org.ID
	db := dbs.DB()
	org = &model.Organization{Model: model.Model{ID: orgID}}
	if err = db.Preload("Members.User").Take(org).Error; err != nil {
		t.Fatal(err)
	}
	members := org.Members
	if len(members) != 1 {
		t.Fatal(members)
	}
	member := members[0]
	if member.Role != model.Owner {
		t.Fatal(member)
	}
	if member.User.Username != "admin" {
		t.Fatal(member)
	}
}
