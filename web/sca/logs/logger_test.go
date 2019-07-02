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
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var testhook = test.NewLocal(defaultLogger)

func TestStderrEqual(t *testing.T) {
	logger := logrus.New()

	if out := logger.Out; out != os.Stderr {
		t.Fatal(out, os.Stderr)
	}

	if out := logrus.StandardLogger().Out; out != os.Stderr {
		t.Fatal(out, os.Stderr)
	}
}
