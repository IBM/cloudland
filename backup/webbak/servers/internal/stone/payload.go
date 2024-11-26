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
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

type Payload struct {
	Zen        string     `json:"zen,omitempty"`
	Ref        string     `json:"ref,omitempty"`
	RefType    string     `json:"ref_type,omitempty"`
	Repository Repository `json:"repository,omitempty"`
}

type Repository struct {
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name,omitempty"`
	CloneUrl      string `json:"clone_url,omitempty"`
}

func (p *Payload) Load(b []byte) (err error) {
	err = json.Unmarshal(b, p)
	return
}

func (p *Payload) Dump() (b []byte, err error) {
	b, err = json.MarshalIndent(p, "", "  ")
	return
}

func (p *Payload) String() (s string) {
	if b, err := p.Dump(); err == nil {
		s = string(b)
	}
	return
}

func Signature(seed string, data []byte) (sum string) {
	mac := hmac.New(sha1.New, []byte(seed))
	if n, err := mac.Write([]byte(data)); err == nil && n == len(data) {
		sum = fmt.Sprintf("sha1=%x", mac.Sum(nil))
	}
	return
}
