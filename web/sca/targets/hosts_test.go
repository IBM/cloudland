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
	"testing"

	"github.com/IBM/cloudland/web/sca/pkgs"
)

func TestFindGroups(t *testing.T) {
	tcs := []struct {
		content string
		want    []string
	}{
		{`---
- name: ensure jaeger installed
  hosts: jaeger
  roles:
    - jaeger
`, []string{"jaeger"}},
		{`---
- name: ensure sidecar installed
  hosts:
    - hyper
    - sci-fe
  roles:
    - sidecar
`, []string{"hyper", "sci-fe"}},
		{`---
- name: ensure cloudland fe installed
  hosts: [sci-fe, hyper]
  roles:
    - scid
`, []string{"sci-fe", "hyper"}},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			content := tc.content
			want := findGroups(content)
			if !pkgs.Strings(want).Equal(pkgs.Strings(tc.want)) {
				t.Fatal(want, tc.want)
			}
		})
	}
}

