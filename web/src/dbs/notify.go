/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import "github.com/lib/pq"

func init() {
}

func eventCallback(eType pq.ListenerEventType, err error) {
	switch eType {
	case pq.ListenerEventConnected:
	case pq.ListenerEventDisconnected:
	case pq.ListenerEventReconnected:
	case pq.ListenerEventConnectionAttemptFailed:
	}
}

func StartListen() {
	//pq.NewListener()
}
