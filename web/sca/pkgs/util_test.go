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
	"os"
	"testing"
)

func TestPkgSum(t *testing.T) {
	contents := PkgSum("cladmin", "v1.0.0", "55ca6286e3e4f4fba5d0448333fa99fc5a404a73", false)
	if contents != `[cladmin]
version = "v1.0.0"
sha1sum = "55ca6286e3e4f4fba5d0448333fa99fc5a404a73"
deploy = false
` {
		t.Fatal(contents)
	}

}

func TestReadDirNames(t *testing.T) {
	tcs := []struct {
		dirname string
		want    func(l int) bool
	}{
		{".", func(l int) bool { return l == 0 }},
		{"..", func(l int) bool { return l > 0 }},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			names := ReadDirNames(tc.dirname)
			if !tc.want(len(names)) {
				t.Fatal(names)
			}
		})
	}
}

func TestCname(t *testing.T) {
	if cname := Cname(); len(cname) != 4 {
		t.Fatal(cname)
	}
}

func TestReadEnvironsNames(t *testing.T) {
	f, err := os.Open("../deploy.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := readEnvironNames(f)
	if len(names) == 0 {
		t.Fatal(names)
	}
}

func TestReadEnvirons(t *testing.T) {
	f, err := os.Open("../deploy.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := []string{}
	environ := readEnvirons(f)
	for name := range environ {
		names = append(names, name)
	}
	ss := Strings(names)
	if !ss.Equal([]string{
		"RELEASE_VERSION",
		"RELEASE_NAME",
		"CLADMIN_PWD",
		"CLADMIN_PID",
		"CLADMIN_ADMIN_LISTEN"}) {
		t.Fatal(names, ss)
	}

}
