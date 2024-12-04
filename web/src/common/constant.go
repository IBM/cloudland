/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package common

type PowerAction string
type SubnetType string

const (
	Stop        PowerAction = "stop"
	HardStop    PowerAction = "hard_stop"
	Start       PowerAction = "start"
	Restart     PowerAction = "restart"
	HardRestart PowerAction = "hard_restart"
	Pause       PowerAction = "pause"
	Resume      PowerAction = "resume"

	Public   SubnetType = "public"
	Internal SubnetType = "internal"
)

var SignedSeret = []byte("Red B")

