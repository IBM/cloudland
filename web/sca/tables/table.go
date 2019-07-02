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
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Table interface {
	Render()
	SetHeader(keys []string)
	Append(row []string)
	AppendBulk(rows [][]string)
	NumLines() int
	SetColWidth(width int)
}

func NewTable(w ...io.Writer) Table {
	writer := io.Writer(os.Stdout)
	if len(w) > 0 {
		writer = w[0]
	}
	table := tablewriter.NewWriter(writer)
	table.SetColWidth(80)
	return table
}
