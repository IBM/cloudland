/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

import "testing"

func TestGetDeployPkg(t *testing.T) {
	tcs := []struct {
		content string
		want    bool
	}{{`[deploy]
version = "v1.0.0"
sha1sum = "55ca6286e3e4f4fba5d0448333fa99fc5a404a73"
`, true},
		{`[cladmin-deploy]
version = "v1.0.0"
sha1sum = "55ca6286e3e4f4fba5d0448333fa99fc5a404a73"
deploy = false
`, false},
		{`[cladmin-deploy]
version = "v1.0.0"
sha1sum = "55ca6286e3e4f4fba5d0448333fa99fc5a404a73"
deploy = true
`, true},
	}
	for _, tc := range tcs {
		tr := LoadContentR(tc.content)
		if tc.want != tr.GetDeploy() {
			t.Fatal(tr)
		}
	}
}

func TestGetInt(t *testing.T) {
	tcs := []struct {
		content string
		want    int
	}{{`[cladmin]
release = "cladmin"
version = "latest"
state = 4`, 4},
{`[cladmin]
release = "cladmin"
version = "latest"
state = 0`, 0},
{`[cladmin]
release = "cladmin"
version = "latest"`, 0},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			tr := LoadContentR(tc.content)
			if state := tr.GetInt("cladmin", "state"); state != tc.want {
				t.Fatal(state, tr.getValue([]string{"cladmin", "state"}).(int64))
			}
		})
	}
}
