/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package targets

import (
	"crypto/sha1"
	fmt "fmt"
	"io/ioutil"
	"os"
)

var (
	KeyErrorNoName = fmt.Errorf("no name specified")
)

func (key *Key) Load(target, name string) (err error) {
	if key.Name == "" {
		key.Name = name
	}
	if name == "" {
		err = KeyErrorNoName
		return
	}

	filename := fmt.Sprintf("targets/%s/keys/%s", target, name)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	content := fmt.Sprintf("%x", sha1.Sum(b))
	key.Private = content
	return
}

func (key *Key) Save(target string) (err error) {
	name := key.GetName()
	if name == "" {
		err = KeyErrorNoName
		return
	}
	dirname := fmt.Sprintf("targets/%s/keys", target)
	err = os.MkdirAll(dirname, 0755)
	if err != nil {
		return
	}
	filename := fmt.Sprintf("%s/%s", dirname, name)
	err = ioutil.WriteFile(filename, []byte(key.Private), 0600)
	return
}

func (key *Key) Remove(target string) (err error) {
	name := key.GetName()
	if name == "" {
		err = KeyErrorNoName
		return
	}
	filename := fmt.Sprintf("targets/%s/keys/%s", target, name)
	err = os.RemoveAll(filename)
	return
}
