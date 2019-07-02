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
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/cloudland/web/sca/pkgs"
	"github.com/spf13/viper"
)

var (
	payloads      = make(chan *Payload, 1024)
	handlePayload = build
)

func init() {
	go handlePayloads()
}

func build(name, version, url string) {
	project := pkgs.Project{Name: name, Version: version}
	project.Build(url)
}

func handlePayloads() {
	for payload := range payloads {
		repo := payload.Repository
		name := repo.FullName
		url := repo.CloneUrl
		url = strings.TrimSuffix(url, ".git")
		url = strings.TrimSuffix(url, name)
		url = strings.TrimSuffix(url, "/")
		tag := payload.Ref
		handlePayload(name, tag, url)
	}
}

func HandleWebHook(w http.ResponseWriter, r *http.Request) {
	secrets := viper.GetStringSlice("webhook.secrets")
	readHeaders := func() (et, sig, id string) {
		header := r.Header
		et = header.Get("X-GitHub-Event")
		sig = r.Header.Get("X-Hub-Signature")
		id = r.Header.Get("X-GitHub-Delivery")
		return
	}
	et, sig, id := readHeaders()
	defer r.Body.Close()
	if et == "" || sig == "" || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		write(w, []byte("{}"))
		return
	}

	var b []byte
	var err error
	if b, err = ioutil.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		write(w, []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	if err = checkSignature(sig, secrets, b); err != nil {
		w.WriteHeader(http.StatusForbidden)
		write(w, b)
		return
	}
	payload := &Payload{}
	if err = payload.Load(b); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		write(w, b)
	}
	w.WriteHeader(http.StatusAccepted)
	if payload.Zen == "" && payload.RefType == "tag" {
		payloads <- payload
	}
	w.Write(b)
}

func checkSignature(signature string, secrets []string,
	b []byte) (err error) {
	for _, secrets := range secrets {
		sum := Signature(secrets, b)
		if sum == signature {
			return
		}
	}
	err = fmt.Errorf("signature does not match.")
	return
}

func write(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
