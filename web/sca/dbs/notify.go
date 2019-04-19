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
