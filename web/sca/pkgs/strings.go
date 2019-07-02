/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

type Strings []string

func (ss Strings) Contains(want string) bool {
	for _, s := range ss {
		if want == s {
			return true
		}
	}
	return false
}

func (ss Strings) Equal(wants Strings) bool {
	for _, s := range ss {
		if !wants.Contains(s) {
			return false
		}
	}
	for _, want := range wants {
		if !ss.Contains(want) {
			return false
		}
	}
	return true
}

func (ss Strings) Append(strs ...string) Strings {
	for _, str := range strs {
		if ss.Contains(str) {
			continue
		}
		ss = Strings(append(ss, str))
	}
	return ss
}
