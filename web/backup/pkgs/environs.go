/*
Copyright <holder> All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

History:
   Date     Who ID    Description
   -------- --- ---   -----------
   01/13/19 nanjj  Initial code

*/

package pkgs

import (
	fmt "fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func LoadEnvirons(target string) (environs map[string]string) {
	environs = map[string]string{}
	tr := LoadTarget(target)
	if tr.NotFound() {
		return
	}
	release := tr.GetString(target, "release")
	version := tr.GetString(target, "version")
	deployf, err := os.Open(fmt.Sprintf("targets/%s/deploy.sh", target))
	if err != nil {
		return
	}
	defer deployf.Close()
	names := readEnvironNames(deployf)
	filename := fmt.Sprintf("targets/%s/environs", target)
	environf, err := os.Open(filename)
	if err != nil {
		environf, err = os.Create(filename)
	}
	if err != nil {
		return
	}
	defer environf.Close()
	environs = readEnvirons(environf)
	for _, name := range names {
		if _, ok := environs[name]; !ok {
			environs[name] = ""
		}
	}
	environs["CLADMIN_PID"] = Pid
	environs["RELEASE_NAME"] = release
	environs["RELEASE_VERSION"] = version
	environs["RELEASE_URL"] = ReleaseURL(release, version)
	return
}

func GetLocalIp() (localIp string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			localIp = ip.String()
			return
		}
	}
	return
}

func ReleaseURL(name, version string) (url string) {
	endpoint := viper.GetString("admin.endpoint")
	if endpoint == "" || strings.Contains(endpoint, "127.0.") {
		endpoint = viper.GetString("admin.listen")
	}
	items := strings.Split(endpoint, ":")
	localIp := ""
	port := "50080"
	if len(items) == 2 {
		localIp = items[0]
		port = items[1]
	}
	if len(localIp) < 4 {
		localIp = GetLocalIp()
	}
	endpoint = fmt.Sprintf("%s:%s", localIp, port)
	url = fmt.Sprintf("http://%s/releases/%s/%s", endpoint, name, version)
	return
}

func SaveEnvirons(target string, environs map[string]string) (err error) {
	filename := fmt.Sprintf("targets/%s/environs", target)
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	lines := []string{}
	for name, value := range environs {
		line := fmt.Sprintf(`export %s=%s`, name, value)
		lines = append(lines, line)
	}
	content := strings.Join(lines, "\n")
	_, err = f.Write([]byte(content))
	return
}
