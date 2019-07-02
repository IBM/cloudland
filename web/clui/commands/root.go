/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package commands

import (
	"context"
	"os"

	"github.com/IBM/cloudland/web/clui/grpcs"
	"github.com/IBM/cloudland/web/clui/routes"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	clientCmd = &cobra.Command{
		Use: "hypercube",
	}
)

func RunDaemon(cmd *cobra.Command, args []string) (err error) {
	g, _ := errgroup.WithContext(context.Background())
	g.Go(routes.Run)
	g.Go(routes.RunRest)
	g.Go(grpcs.Run)
	return g.Wait()
}

func RootCmd() (cmd *cobra.Command) {
	for _, arg := range os.Args {
		if arg == "--daemon" {
			daemonCmd := &cobra.Command{
				Use:  "hypercube",
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
