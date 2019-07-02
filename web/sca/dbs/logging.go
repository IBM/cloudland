/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package dbs

import (
	context "context"

	"github.com/IBM/cloudland/web/sca/logs"
)

var (
	tracing *logs.Tracing
)

func getTracing() *logs.Tracing {
	if tracing == nil {
		tracing = logs.StartTracing("DBAdmin")
	}
	return tracing
}

func startLogging(ctx context.Context, name string) (*logs.SpanLogger, context.Context) {
	return getTracing().StartLogging(ctx, name)
}
