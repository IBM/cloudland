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
	"os"
	"strings"

	toml "github.com/pelletier/go-toml"
)

type TreeR struct {
	*toml.Tree
	filename string
}

func (tr *TreeR) GetString(name ...string) (value string) {
	v := tr.getValue(name)
	if v != nil {
		value = v.(string)
	}
	return
}

func (tr *TreeR) GetBool(name ...string) (value bool) {
	v := tr.getValue(name)
	if v != nil {
		value, _ = v.(bool)
	}
	return
}

func (tr *TreeR) GetInt(name ...string) (value int) {
	v := tr.getValue(name)
	if v != nil {
		if n, ok := v.(int64); ok {
			value = int(n)
		}
	}
	return
}

func (tr *TreeR) getValue(name []string) (value interface{}) {
	if tr.Tree == nil {
		return
	}
	value = tr.GetPath(name)
	return
}

func LoadPkgSum(name, version string) (tr *TreeR) {
	f := fmt.Sprintf("packages/%s/%s/%s.sum", name, version, name)
	tr = LoadFileR(f)
	return
}

func LoadFileR(p string) (tr *TreeR) {
	tr = &TreeR{
		filename: p,
	}
	if tree, err := toml.LoadFile(p); err == nil {
		tr.Tree = tree
	}
	return
}

func LoadContentR(content string) (tr *TreeR) {
	tr = &TreeR{}
	if tree, err := toml.Load(content); err == nil {
		tr.Tree = tree
	}
	return
}

func LoadRelease(name, version string) (tr *TreeR) {
	f := fmt.Sprintf("releases/%s/%s/release.toml", name, version)
	tr = LoadFileR(f)
	return
}

func LoadTarget(target string) (tr *TreeR) {
	f := fmt.Sprintf("targets/%s/target.toml", target)
	tr = LoadFileR(f)
	return
}

func LoadReleasePkg(name string, pkg string) (tr *TreeR) {
	f := fmt.Sprintf("release/%s/%s/%s.sum", name, "latest", pkg)
	tr = LoadFileR(f)
	return
}

func (tr *TreeR) NotFound() bool {
	return tr.Tree == nil
}

func (tr *TreeR) Remove() (err error) {
	if tr.NotFound() {
		err = fmt.Errorf("component not found")
		return
	}
	filename := tr.filename
	if filename == "" {
		return
	}
	idx := strings.LastIndex(filename, "/")
	if idx == -1 {
		err = fmt.Errorf("directory not found")
		return
	}
	dirname := filename[0:idx]
	fi, err := os.Stat(dirname)
	if err != nil {
		return
	}
	if !fi.IsDir() {
		err = fmt.Errorf("%s is not directory", dirname)
		return
	}
	err = os.RemoveAll(dirname)
	return
}

func (tr *TreeR) GetDeploy() (deploy bool) {
	for _, key := range tr.Keys() {
		pkg := tr.GetPkg(key)
		if pkg.GetDeploy() {
			deploy = true
			return
		}
	}
	return
}

func (tr *TreeR) GetPkg(name string) (pkg *Pkg) {
	version := tr.GetString(name, "version")
	sha1sum := tr.GetString(name, "sha1sum")
	deploy := tr.GetBool(name, "deploy")
	if !deploy {
		deploy = name == "deploy"
	}
	if version != "" && sha1sum != "" {
		pkg = &Pkg{
			Name:    name,
			Version: version,
			Sha1Sum: sha1sum,
			Deploy:  deploy,
		}
	}
	return
}

func (tr *TreeR) GetPkgs() (ps []*Pkg) {
	ps = []*Pkg{}
	for _, key := range tr.Keys() {
		p := tr.GetPkg(key)
		if p == nil {
			continue
		}
		ps = append(ps, p)
	}
	return
}
