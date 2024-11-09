/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package stone

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestHandleWebhook(t *testing.T) {
	defer func() {
		handlePayload = build
	}()
	results := make(chan string, 3)
	handlePayload = func(name, version, url string) {
		results <- name
		results <- version
		results <- url
	}
	b, err := ioutil.ReadFile("payload.json")
	if err != nil {
		t.Fatal(err)
	}
	secret := "abc123"
	viper.SetDefault("webhook.secrets", []string{secret})
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://example.com",
		strings.NewReader(string(b)))
	if err != nil {
		t.Fatal()
	}
	header := req.Header
	signature := Signature(secret, b)
	header.Set("X-GitHub-Event", "create")
	header.Set("X-Hub-Signature", signature)
	header.Set("X-GitHub-Delivery", "3efc8b70-b123-11e8-8ab0-6e043d6feeae")
	HandleWebHook(w, req)
	if w.Code != 202 {
		t.Fatal(w.Body)
	}

	if name := <-results; name != "nanjj/ncatd" {
		t.Fatal(name)
	}
	if version := <-results; version != "v1.0.1" {
		t.Fatal(version)
	}
	if url := <-results; url != "https://github.com" {
		t.Fatal(url)
	}

}
