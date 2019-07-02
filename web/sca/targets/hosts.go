/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package targets

import (
	fmt "fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/cloudland/web/sca/pkgs"
)

var (
	HostErrorFormat     = fmt.Errorf(`should in format [<group>]\n<hostname> ansible_host=<ip>`)
	HostErrorNotSame    = fmt.Errorf("hostname not same with inventory file name")
	HostErrorNoIp       = fmt.Errorf("no ip address")
	HostErrorNoRole     = fmt.Errorf("no group")
	HostErrorNoHostname = fmt.Errorf("no hostname")
	HostErrorReserved   = fmt.Errorf("hostname reservced")
	HostnameReserved    = pkgs.Strings([]string{
		"0local", "all",
	})
)

func (host *Host) Save(target string) (err error) {
	dirname := fmt.Sprintf("targets/%s/hosts", target)
	err = os.MkdirAll(dirname, 0755)
	if err != nil {
		return
	}
	name := host.Name
	if name == "" {
		err = HostErrorNoHostname
		return
	}
	if HostnameReserved.Contains(name) {
		err = HostErrorReserved
		return
	}
	filename := fmt.Sprintf("%s/%s", dirname, name)
	err = ioutil.WriteFile(filename, []byte(fmt.Sprintf(`[%s]
%s ansible_host=%s`, host.GetGroup(), host.GetName(), host.GetIp())), 0644)
	return
}

func (host *Host) Load(target string, name string) (err error) {
	if host.Name == "" {
		host.Name = name
	}
	filename := fmt.Sprintf("targets/%s/hosts/%s", target, name)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	s := strings.TrimSpace(string(b))
	err = host.load(s)
	return
}

func (host *Host) load(content string) (err error) {
	items := strings.Split(content, "\n")
	if len(items) != 2 {
		err = HostErrorFormat
		return
	}
	group := strings.Trim(items[0], "[]")
	if group == "" {
		err = HostErrorNoRole
		return
	}
	idx := strings.Index(items[1], "ansible_host=")
	if idx != -1 {
		line := items[1]
		hostname := strings.TrimSpace(line[0:idx])
		if hostname == "" {
			err = HostErrorNoHostname
			return
		}
		if host.Name != "" && hostname != host.Name {
			err = HostErrorNotSame
			return
		}
		ip := strings.TrimSpace(line[idx+len("ansible_host="):])
		if ip == "" {
			err = HostErrorNoIp
			return
		}
		host.Name = hostname
		host.Ip = ip
		host.Group = group
	}
	return
}

func LoadHostnames(target string) (hostnames []string) {
	names := pkgs.ReadFileNames(fmt.Sprintf("targets/%s/hosts", target))
	for _, name := range names {
		if !HostnameReserved.Contains(name) {
			hostnames = append(hostnames, name)
		}
	}
	return
}

func FindGroups(target string) (groups []string) {
	dirname := fmt.Sprintf("targets/%s", target)
	names := pkgs.ReadFileNames(dirname)
	if len(names) == 0 {
		return
	}
	result := pkgs.Strings([]string{})
	for _, name := range names {
		filename := fmt.Sprintf("%s/%s", dirname, name)
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}
		content := string(b)
		gs := findGroups(content)
		for _, g := range gs {
			if !HostnameReserved.Contains(g) &&
				!result.Contains(g) {
				result = result.Append(g)
			}
		}
	}
	return result
}

func findGroups(content string) (groups []string) {
	lines := strings.Split(content, "\n")
	hostFound := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if hostFound {
			if line == "" {
				continue
			}
			if line[0] == '-' {
				line = strings.Trim(line, `- "`)
				groups = append(groups, line)
			} else {
				break
			}
		}
		if !strings.HasPrefix(line, "hosts:") {
			continue
		}
		if hosts := line[len("hosts:"):]; hosts != "" {
			hosts = strings.Trim(strings.TrimSpace(hosts), "[]")
			if hosts == "" {
				continue
			}
			items := strings.Split(hosts, ",")
			if len(items) == 1 {
				items = strings.Split(hosts, ":")
			}
			for _, item := range items {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				groups = append(groups, item)
			}
			break
		} else {
			hostFound = true
		}
	}
	return
}
