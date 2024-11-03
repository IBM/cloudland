/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package rpcs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

type ExecuteRequest struct {
	Id      int32
	Extra   int32
	Control string
	Command string
}

type ExecuteReply struct {
	Status string
}

var remoteExecPath string


func HyperExecute(ctx context.Context, control, command string) (err error) {
	execReq := &ExecuteRequest{
		Id: 100,
		Extra: 0,
		Control: control,
		Command: command,
	}
	jsonReq, err := json.Marshal(execReq)
	payload := bytes.NewBufferString(string(jsonReq))
	if remoteExecPath == "" {
		remoteExecPath = viper.GetString("sci.endpoint") + "/internal/execute"
	}
	log.Printf("remotePath: %s, jsonPayload: %v", remoteExecPath, payload)
	resp, err := http.Post(remoteExecPath, "application/json", payload)
	if err != nil {
		log.Println("Error posting data:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	log.Println("Response Status:", resp.Status)
	log.Println("Response Body:", string(body))
	return
}
