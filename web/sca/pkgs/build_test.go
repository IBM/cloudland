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
	"os/exec"
	"testing"
)

func TestBuild(t *testing.T) {
}

func TestDetach(t *testing.T) {
	t.Skip()
	cmd := exec.Command("nohup", "sleep", "10")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Process.Release()
}
