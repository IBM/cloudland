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
	"context"
	"log"

	"github.com/IBM/cloudland/web/clui/jobs"
	"github.com/IBM/cloudland/web/clui/scripts"
	"github.com/IBM/cloudland/web/sca/clients"
)

func RemoteExecClient() scripts.RemoteExecClient {
	return scripts.NewRemoteExecClient(clients.GetClientConn("sci"))
}

func JobServiceClient(endpoint string) jobs.JobServiceClient {
	cc := clients.GetClientConn(endpoint)
	return jobs.NewJobServiceClient(cc)
}

func HyperExecute(ctx context.Context, control, command string) (err error) {
	if control == "" {
		return
	}
	sciClient := RemoteExecClient()
	sciReq := &scripts.ExecuteRequest{
		Id:      100,
		Extra:   0,
		Control: control,
		Command: command,
	}
	_, err = sciClient.Execute(ctx, sciReq)
	if err != nil {
		log.Println("SCI client execution failed, %v", err)
		return
	}
	return
}
