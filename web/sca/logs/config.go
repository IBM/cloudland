/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	"strings"

	"github.com/IBM/cloudland/web/sca/configs"
	"github.com/spf13/viper"
)

var (
	config = &configs.Config{viper.GetViper()}
)

func Config(v *viper.Viper) {
	config.Viper = v
}

func GetJaegerURI() string {
	return config.GetString("127.0.0.1:6381", "jaeger.endpoint", "jaeger.url", "jaeger.uri")
}

func GetListenAddr() string {
	jaegerURI := GetJaegerURI()
	idx := strings.Index(jaegerURI, ":")
	if idx != -1 {
		jaegerURI = jaegerURI[idx:]
	}
	return jaegerURI
}
