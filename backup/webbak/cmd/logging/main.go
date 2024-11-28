/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package main

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{}
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		RootCmd.Println(err)
	}
}
