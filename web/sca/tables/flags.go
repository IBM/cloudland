/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package tables

import (
	"github.com/spf13/cobra"
)

type Options struct {
	*cobra.Command
}

func (cmd *Options) GetInt(name string) (i int) {
	var err error
	if i, err = cmd.Flags().GetInt(name); err != nil {
		panic(err)
	}
	return
}

func (cmd *Options) GetInt32(name string) (i int32) {
	var err error
	if i, err = cmd.Flags().GetInt32(name); err != nil {
		panic(err)
	}
	return
}

func (cmd *Options) GetString(name string) (s string) {
	var err error
	if s, err = cmd.Flags().GetString(name); err != nil {
		panic(err)
	}
	return s
}

func (cmd *Options) GetStringSlice(name string) (s []string) {
	var err error
	if s, err = cmd.Flags().GetStringSlice(name); err != nil {
		panic(err)
	}
	return s
}

func (cmd *Options) GetBool(name string) (b bool) {
	var err error
	if b, err = cmd.Flags().GetBool(name); err != nil {
		panic(err)
	}
	return b
}
