/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

import (
	fmt "fmt"
	"strings"

	"github.com/coduno/netrc"
)

// AddNetrc item in format github.com=login:pass
func AddNetrc(token string) (err error) {
	github, login, pass, err := ParseNetrcToken(token)
	if err != nil {
		return
	}
	entries, err := netrc.Parse()
	if err != nil {
		entries = netrc.Entries{}
		err = nil
	}
	entries[github] = netrc.Entry{
		Login:    login,
		Password: pass,
	}
	err = entries.Save()
	return
}

func RemoveNetrc(github string) (err error) {

	entries, err := netrc.Parse()
	if err != nil {
		err = nil
		return
	}

	if _, ok := entries[github]; ok {
		delete(entries, github)
		err = entries.Save()
		return
	}

	return
}

func ParseNetrcToken(token string) (github, login, pass string, err error) {
	token = strings.TrimSpace(token)
	items := strings.Split(token, "=")
	switch len(items) > 1 {
	case true:
		github = strings.TrimSpace(items[0])
		login = strings.TrimSpace(items[1])
	case false:
		github = "github.com"
		login = strings.TrimSpace(items[0])
	}
	items = strings.Split(login, ":")
	if len(items) == 1 {
		err = fmt.Errorf("not in format github.com=login:pass")
		return
	}
	login = items[0]
	pass = items[1]
	return
}

func NetrcEntries() (entries []string) {
	items, err := netrc.Parse()
	if err != nil {
		return
	}
	for host, item := range items {
		entry := fmt.Sprintf("%s=%s:<pass>", host, item.Login)
		entries = append(entries, entry)
	}
	return
}
