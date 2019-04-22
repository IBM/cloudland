/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package grpcs

import (
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
