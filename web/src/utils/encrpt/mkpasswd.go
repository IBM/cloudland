/*
Copyright PEG Tech Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package encrpt

import (
	"fmt"

	"github.com/tredoe/osutil/user/crypt"
	"github.com/tredoe/osutil/user/crypt/apr1_crypt"
	"github.com/tredoe/osutil/user/crypt/common"
	"github.com/tredoe/osutil/user/crypt/md5_crypt"
	"github.com/tredoe/osutil/user/crypt/sha256_crypt"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// Mkpasswd
// Generate a password hash using the specified hash type.
// The hash types supported are: md5, sha256, sha512, and apr1.
func Mkpasswd(plain_text_pw string, hash_type string) (string, error) {
	var crypter crypt.Crypter
	var s common.Salt // salt parameters related to c
	var err error
	switch hash_type {
	case "md5":
		crypter = md5_crypt.New()
		s = md5_crypt.GetSalt()
	case "sha256":
		crypter = sha256_crypt.New()
		s = sha256_crypt.GetSalt()
	case "sha512":
		crypter = sha512_crypt.New()
		s = sha512_crypt.GetSalt()
	case "apr1":
		crypter = apr1_crypt.New()
		s = apr1_crypt.GetSalt()
	default:
		return "", fmt.Errorf("Unknown hash type: %s", hash_type)
	}
	if err != nil {
		return "", err
	}
	saltString := string(s.GenerateWRounds(s.SaltLenMax, 4096))
	//return "$6$rounds=4096$TfbcMlkfjT91t0Xk$g76iQZz8RJJ.CmsL4sYmxGN7SnOpXrh8e8vTHdRXgozf7HMl2DBgkHvE7Jp6AXIbLHbo9vgEHqdyxmGYhnCiX0", nil
	return crypter.Generate([]byte(plain_text_pw), []byte(saltString))
}
