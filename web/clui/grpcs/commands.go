/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package grpcs

import (
	"fmt"
	"strings"
)

func EncodeCommand(name string, args ...interface{}) (command string) {
	command = name
	l := len(args)
	if l == 0 {
		return
	}
	for i := 0; i < l; i++ {
		command = fmt.Sprintf("%s '%v'", command, args[i])
	}
	return
}

func popFirst(content string) (first, left string) {
	const cutset = " '"
	idx := strings.Index(content, cutset)
	if idx != -1 {
		first = strings.Trim(strings.TrimSpace(content[0:idx]), "'")

		left = strings.Trim(strings.TrimSpace(content[idx:]), "'")
	} else {
		first = strings.Trim(strings.TrimSpace(content), "'")
	}
	return
}

func parseCommand(s string) (cmd string) {
	idx := strings.LastIndex(s, "/")
	if idx != -1 {
		s = strings.TrimPrefix(s[idx:], "/")
	}
	cmd = strings.TrimSuffix(s, ".sh")
	return
}

func DecodeCommand(content string) (cmd string, args []string) {
	if content == "" {
		return
	}
	first, left := popFirst(content)
	cmd = parseCommand(first)
	args = append(args, first)
	for left != "" {
		first, left = popFirst(left)
		args = append(args, first)
	}
	return
}

func AppendAttachment(command string, attachment string) (content string) {
	content = fmt.Sprintf("%s <<EOF\n%s\nEOF\n", command, attachment)
	return
}
