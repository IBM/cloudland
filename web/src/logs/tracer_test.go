/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package logs

import (
	context "context"
	"testing"
)

func TestStartSpanFromContext(t *testing.T) {
}

func TestNewTracer(t *testing.T) {
	tr := StartTracing("TestNewTracer")
	if tr == nil {
		t.Fatal()
	}
	logger, ctx := tr.StartLogging(context.Background(), "TestIt")
	if logger == nil {
		t.Fatal()
	}
	defer logger.Finish()
	logger.Error("hi")
	t.Log(ctx)
}
