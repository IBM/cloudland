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
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	fmt "fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	Executable, _ = os.Executable()
	Pid           = fmt.Sprintf("%v", os.Getpid())
)

func Extract() (filenames []string, err error) {
	src := Executable
	dest, err := os.Getwd()
	if err != nil {
		return
	}
	filenames, err = Unzip(src, dest)
	return
}

func OpenZipFile(src string) (r *bytes.Reader, fs int64, err error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	archiveStart := readArchiveStart(f)
	if archiveStart < 0 {
		return nil, 0, errors.New("Invalid format zip")
	}

	_, err = f.Seek(archiveStart, 0)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	r = bytes.NewReader(b)
	fs = int64(r.Size())
	return
}

func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	f, fs, err := OpenZipFile(src)
	if err != nil {
		return filenames, err
	}
	r, err := zip.NewReader(f, fs)
	if err != nil {
		return filenames, err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			_, err = io.Copy(outFile, rc)
			// Close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return filenames, err
			}
		}
	}
	return filenames, nil
}

func readArchiveStart(f *os.File) (startOfArchive int64) {
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return -1
	}
	size := fi.Size()

	// look for directoryEndSignature in the last 1k, then in the last 65k
	var buf []byte
	var directoryEndOffset int64
	findSignatureInBlock := func(b []byte) int64 {
		for i := len(b) - 22; i >= 0; i-- {
			// defined from directoryEndSignature in struct.go
			if b[i] == 'P' && b[i+1] == 'K' && b[i+2] == 0x05 && b[i+3] == 0x06 {
				// n is length of comment
				n := int(b[i+22-2]) | int(b[i+22-1])<<8
				if n+22+i <= len(b) {
					return int64(i)
				}
			}
		}
		return -1
	}
	for i, bLen := range []int64{1024, 65 * 1024} {
		if bLen > size {
			bLen = size
		}
		buf = make([]byte, int(bLen))
		if _, err := f.ReadAt(buf, size-bLen); err != nil && err != io.EOF {
			return -1
		}
		if p := findSignatureInBlock(buf); p >= 0 {
			buf = buf[p:]
			directoryEndOffset = size - bLen + p
			break
		}
		if i == 1 || bLen == size {
			return -1
		}
	}

	b := buf[4+2+2+2+2:] // skip signature
	readUint32 := func() int64 {
		v := binary.LittleEndian.Uint32(b)
		b = b[4:]
		return int64(v)
	}

	directorySize := readUint32()
	directoryOffset := readUint32()
	// Calculate where the zip data actually begins
	startOfArchive = directoryEndOffset - directorySize - directoryOffset
	return
}
