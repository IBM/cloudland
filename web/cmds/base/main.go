/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/viper"

	"web/src/routes"
	"web/src/rpcs"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	Version = "0.0.1"
)

var (
	clientCmd = &cobra.Command{
		Use: "clbase",
	}
)

func RunDaemon(cmd *cobra.Command, args []string) (err error) {
	g, _ := errgroup.WithContext(context.Background())
	g.Go(routes.Run)
	g.Go(rpcs.Run)
	return g.Wait()
}

func RootCmd() (cmd *cobra.Command) {
	for _, arg := range os.Args {
		if arg == "--daemon" {
			daemonCmd := &cobra.Command{
				Use:  "CloudlandBase",
				RunE: RunDaemon,
			}
			daemonCmd.Flags().Bool("daemon", false, "daemon")
			viper.SetConfigFile("conf/config.toml")
			if err := viper.ReadInConfig(); err != nil {
				daemonCmd.Println(err)
				return
			}
			return daemonCmd
		}
	}
	return clientCmd
}
func init() {
	viper.Set("AppVersion", Version)
	viper.Set("GoVersion", strings.Title(runtime.Version()))
	file := "/opt/cloudland/log/clbase.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetPrefix("[webUILog]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func main() {
	rootCmd := RootCmd()
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
	}
}
