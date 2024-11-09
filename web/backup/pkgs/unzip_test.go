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

func TestUnzip(t *testing.T) {
	filenames, err := Unzip("pkgs_zip_data", "/tmp")
	if err != nil {
		t.Fatal(err, filenames)
	}
}

func TestOpenZipFile(t *testing.T) {
	r, size, err := OpenZipFile("pkgs_zip_data")
	if err != nil {
		t.Fatal(err, size, r)
	}
}

func TestReadArchiveStart(t *testing.T) {
	tcs := []struct {
		src    string
		offset int64
	}{
		{"unzip.go", -1},
		{"pkgs_zip_data", 1110},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			f, err := os.Open(tc.src)
			if err != nil {
				t.Fatal(err, f)
			}
			offset := readArchiveStart(f)
			if offset != tc.offset {
				t.Fatal(offset)
			}
		})
	}
}
