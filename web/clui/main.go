/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"runtime"
	"strings"

	"github.com/IBM/cloudland/web/clui/commands"
	"github.com/spf13/viper"
)

var (
	Version = "0.0.1"
)

func init() {
	viper.Set("AppVersion", Version)
	viper.Set("GoVersion", strings.Title(runtime.Version()))
}

func main() {
	rootCmd := commands.RootCmd()
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
	}
}
