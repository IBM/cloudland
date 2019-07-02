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
	"bufio"
	fmt "fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"sort"

	vers "github.com/hashicorp/go-version"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func PkgSum(name, version, sha1sum string, deploy bool) string {
	if name == "deploy" {
		deploy = true
	}
	return fmt.Sprintf(`[%s]
version = "%s"
sha1sum = "%s"
deploy = %v
`, name, version, sha1sum, deploy)
}

func NameVersion(pkg string) (name, version string) {
	if tokens := strings.Split(pkg, "="); len(tokens) != 2 {
		name = pkg
	} else {
		name = tokens[0]
		version = tokens[1]
	}
	return
}

func Names(ps []*Pkg) string {
	names := []string{}
	for _, p := range ps {
		name := p.GetName()
		version := p.GetVersion()
		deploy := p.GetDeploy()
		if !deploy {
			deploy = name == "deploy"
		}
		if version != "" {
			name += "=" + version
		}
		if deploy {
			name += " ✔"
		}
		names = append(names, name)
	}
	s := strings.Join(names, "\n")
	s = strings.TrimSpace(s)
	if s == "" {
		s = "-"
	}
	return s
}

func LatestVersion(name string) (latest string) {
	raws := Versions(name)
	if len(raws) == 0 {
		return
	}
	versions := []*vers.Version{}
	for _, raw := range raws {
		version, err := vers.NewVersion(raw)
		if err != nil {
			continue
		}
		versions = append(versions, version)
	}
	sort.Sort(vers.Collection(versions))
	if len(versions) == 0 {
		return
	}
	version := versions[len(versions)-1]
	latest = version.Original()
	return
}

func Versions(name string) (versions []string) {
	dirname := fmt.Sprintf("packages/%s", name)
	return ReadDirNames(dirname)
}

func readNames(dirname string, isdir bool) (names []string) {
	names = []string{}
	fis, err := ioutil.ReadDir(dirname)
	if err != nil {
		return
	}
	for _, fi := range fis {
		if isdir {
			if !fi.IsDir() { // skip regular files
				continue
			}
		} else {
			if fi.IsDir() { // skip directory files
				continue
			}
		}
		name := fi.Name()
		if name[0] == '.' { // skip hiden directories
			continue
		}
		names = append(names, name)
	}
	if len(names) > 0 {
		sort.Strings(names)
	}
	return

}

func ReadFileNames(dirname string) (names []string) {
	return readNames(dirname, false)
}

func ReadDirNames(dirname string) (names []string) {
	return readNames(dirname, true)
}

func Cname() string {
	return fmt.Sprintf("%04d", rand.Int31n(10000))
}

func EnvReserved(name string) bool {
	return Strings([]string{
		"RELEASE_NAME", "RELEASE_VERSION", "RELEASE_URL"}).Contains(name)
}


func readEnvironNames(r io.Reader) (names []string) {
	names = []string{}
	for name := range readEnvirons(r) {
		names = append(names, name)
	}
	return
}

func readEnvirons(r io.Reader) (environ map[string]string) {
	environ = map[string]string{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "export ") {
			line = line[7:]
			idx := strings.IndexByte(line, '=')
			if idx != -1 {
				name := strings.TrimSpace(line[0:idx])
				value := strings.TrimSpace(line[idx+1:])
				value = strings.Trim(value, `'"`)
				environ[name] = value
			}
		}
	}
	return
}
