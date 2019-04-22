/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkgs

import "testing"

func TestLoadEnvirons(t *testing.T) {
	environs := LoadEnvirons("cladmin")
	for k, v := range environs {
		t.Log(k, v)
	}
}
