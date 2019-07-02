/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package stone

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestSignature(t *testing.T) {
	sum := Signature("passw0rd", []byte("hello"))
	if sum != "sha1=c74e41881339bfcae42f9c9cdde2f3a1ee061f7d" {
		t.Fatal(sum)
	}
}

func TestPayload(t *testing.T) {
	b, err := ioutil.ReadFile("payload.json")
	if err != nil {
		t.Fatal(err)
	}
	v := &Payload{}
	err = json.Unmarshal(b, v)
	if err != nil {
		t.Fatal(err)
	}
	if repo := v.Repository; v.RefType != "tag" || v.Ref != "v1.0.1" ||
		repo.CloneUrl != "https://github.com/nanjj/ncatd.git" ||
		repo.FullName != "nanjj/ncatd" || repo.Name != "ncatd" {
		t.Fatal(v)
	}
}
