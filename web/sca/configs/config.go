/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package configs

import "github.com/spf13/viper"

type Config struct {
	*viper.Viper
}

func (v *Config) GetInt(defvalue int, keys ...string) (value int) {
	for _, key := range keys {
		value = v.Viper.GetInt(key)
		if value != 0 {
			return
		}
	}
	return defvalue
}

func (v *Config) GetString(defvalue string, keys ...string) (value string) {
	for _, key := range keys {
		value = v.Viper.GetString(key)
		if value != "" {
			return
		}
	}
	return defvalue
}

func (v *Config) GetBool(defvalue bool, keys ...string) (value bool) {
	for _, key := range keys {
		value = v.Viper.GetBool(key)
		if value != defvalue {
			return
		}
	}
	return defvalue
}
