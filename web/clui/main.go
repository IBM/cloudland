/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/IBM/cloudland/web/clui/commands"
	"github.com/spf13/viper"
	"log"
	"os"
	"runtime"
	"strings"
)

var (
	Version = "0.0.1"
)

func init() {
	viper.Set("AppVersion", Version)
	viper.Set("GoVersion", strings.Title(runtime.Version()))
	file := "/opt/cloudland/log/clui.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetPrefix("[webUILog]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func main() {
	rootCmd := commands.RootCmd()
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
	}
}
