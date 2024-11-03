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
	"os/exec"
	"strings"

	toml "github.com/pelletier/go-toml"
)

func (p *Project) Build(url string) (pkgs []*Pkg, err error) {
	repo := p.GetName()
	version := p.GetVersion()
	// setup
	if _, err = RunParts("setup"); err != nil {
		return
	}
	// build
	output, err := RunParts("build", Cname(), repo, url, version)
	if err != nil {
		err = fmt.Errorf("err: %v, output: %s", err, output)
		return
	}
	lines := []string{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# Output: ") {
			line = line[10:]
			lines = append(lines, line)
		}
	}
	output = strings.Join(lines, "\n")
	if output == "" {
		return
	}
	tree, err := toml.LoadBytes([]byte(output))
	if err != nil {
		return
	}
	names := tree.Keys()
	getString := func(keys ...string) (value string) {
		v := tree.GetPath(keys)
		if v != nil {
			value = v.(string)
		}
		return
	}
	for _, name := range names {
		version := getString(name, "version")
		sha1sum := getString(name, "sha1sum")
		pkg := &Pkg{
			Name:    name,
			Version: version,
			Sha1Sum: sha1sum,
		}
		pkgs = append(pkgs, pkg)
	}

	return
}
func RunParts(part string, args ...string) (output string, err error) {
	return runParts(false, part, args...)
}
func runParts(detach bool, part string, args ...string) (output string, err error) {
	if !strings.HasPrefix(part, "scripts/") {
		part = fmt.Sprintf("scripts/%s", part)
	}
	for i, arg := range args {
		args[i] = fmt.Sprintf("--arg=%s", arg)
	}
	name := "run-parts"
	args = append([]string{part}, args...)
	if detach {
		args = append([]string{name}, args...)
		name = "nohup"
	}
	cmd := exec.Command(name, args...)
	if detach {
		err = cmd.Start()
		if err != nil {
			err = fmt.Errorf("%v, cmd: %s, args: %s", err, cmd.Path,
				strings.Join(cmd.Args, " "))
			return
		}
		err = cmd.Process.Release()
		return
	}
	b, err := cmd.Output()
	if err == nil {
		output = string(b)
	} else {
		err = fmt.Errorf("%v, cmd: %s, args: %s", err, cmd.Path,
			strings.Join(cmd.Args, " "))
	}
	return
}

func RunPartsDetach(part string, args ...string) (err error) {
	_, err = runParts(true, part, args...)
	return
}
