/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	"github.com/IBM/cloudland/web/sca/configs"
	"github.com/spf13/viper"
)

var (
	cfg = &config{
		&configs.Config{Viper: viper.GetViper()},
	}
)

func Config(v *viper.Viper) {
	cfg.Viper = v
}

type config struct{ *configs.Config }

func (v *config) GetIdle() int {
	return v.GetInt(1, "db.idle")
}

func (v *config) GetOpen() int {
	return v.GetInt(50, "db.open")
}

func (v *config) GetLifetime() int {
	return v.GetInt(30, "db.lifetime")
}

func (v *config) GetUri() string {
	return v.GetString("dbs.db", "db.uri", "db.url")
}

func (v *config) GetType() string {
	return v.GetString("sqlite3", "db.type")
}

func (v *config) GetDebug() bool {
	return v.GetBool(false, "db.debug")
}
