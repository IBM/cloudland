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

func TestAppend(t *testing.T) {
	s1 := Strings([]string{"a", "b"})
	s1 = s1.Append("c")
	if len(s1) != 3 {
		t.Fatal(s1)
	}
}

func TestEqual(t *testing.T) {
	tcs := []struct {
		a    Strings
		b    Strings
		want bool
	}{
		{Strings([]string{}), Strings([]string{}), true},
		{Strings([]string{"a"}), Strings([]string{"b"}), false},
		{Strings([]string{"a", "b"}), Strings([]string{"b", "c"}), false},
		{Strings([]string{"b", "c"}), Strings([]string{"b", "c"}), true},
		{Strings([]string{"b", "c"}), Strings([]string{"c", "b"}), true},
		{Strings([]string{"b", "c"}), Strings([]string{"b", "c", "b"}), true},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			if want := tc.a.Equal(tc.b); want != tc.want {
				t.Fatal(tc.a, tc.b, tc.want, want)
			}
		})
	}
}
