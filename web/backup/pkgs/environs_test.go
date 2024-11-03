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

func TestLoadEnvirons(t *testing.T) {
	environs := LoadEnvirons("cladmin")
	for k, v := range environs {
		t.Log(k, v)
	}
}
