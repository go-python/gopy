// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bufio"
	"bytes"
	"fmt"
	"go/types"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func isErrorType(typ types.Type) bool {
	return typ == types.Universe.Lookup("error").Type()
}

func isStringer(obj types.Object) bool {
	switch obj := obj.(type) {
	case *types.Func:
		if obj.Name() != "String" {
			return false
		}
		sig, ok := obj.Type().(*types.Signature)
		if !ok {
			return false
		}
		if sig.Recv() == nil {
			return false
		}
		if sig.Params().Len() != 0 {
			return false
		}
		res := sig.Results()
		if res.Len() != 1 {
			return false
		}
		ret := res.At(0).Type()
		if ret != types.Universe.Lookup("string").Type() {
			return false
		}
		return true
	default:
		return false
	}
}

func hasError(sig *types.Signature) bool {
	res := sig.Results()
	if res == nil || res.Len() <= 0 {
		return false
	}

	nerr := 0
	for i := 0; i < res.Len(); i++ {
		ret := res.At(i)
		if isErrorType(ret.Type()) {
			nerr++
		}
	}

	switch {
	case nerr == 0:
		return false
	case nerr == 1:
		return true
	default:
		panic(fmt.Errorf(
			"gopy: invalid number of comma-errors (%d)",
			nerr,
		))
	}
}

func isConstructor(sig *types.Signature) bool {
	//TODO(sbinet)
	return false
}

// getPkgConfig returns the name of the pkg-config python's pc file
func getPkgConfig(vers int) (string, error) {
	bin, err := exec.LookPath("pkg-config")
	if err != nil {
		return "", fmt.Errorf(
			"gopy: could not locate 'pkg-config' executable (err: %v)",
			err,
		)
	}

	out, err := exec.Command(bin, "--list-all").Output()
	if err != nil {
		return "", fmt.Errorf(
			"gopy: error retrieving the list of packages known to pkg-config (err: %v)",
			err,
		)
	}

	pkgs := []string{}
	re := regexp.MustCompile(fmt.Sprintf(`^python(\s|-|\.|)%d.*?`, vers))
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		err = s.Err()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		line := s.Bytes()
		if !bytes.HasPrefix(line, []byte("python")) {
			continue
		}

		if !re.Match(line) {
			continue
		}

		pkg := bytes.Split(line, []byte(" "))
		pkgs = append(pkgs, string(pkg[0]))
	}

	if err != nil {
		return "", fmt.Errorf(
			"gopy: error scanning pkg-config output (err: %v)",
			err,
		)
	}

	if len(pkgs) <= 0 {
		return "", fmt.Errorf(
			"gopy: could not find pkg-config file (no python.pc installed?)",
		)
	}

	sort.Strings(pkgs)

	// FIXME(sbinet): make sure we take the latest version?
	pkgcfg := pkgs[0]

	return pkgcfg, nil
}

func getGoVersion(version string) (int64, int64, error) {
	version_regex := regexp.MustCompile(`^go((\d+)(\.(\d+))*)`)
	match := version_regex.FindStringSubmatch(version)
	if match == nil {
		return -1, -1, fmt.Errorf("gopy: invalid Go version information: %q", version)
	}
	version_info := strings.Split(match[1], ".")
	major, _ := strconv.ParseInt(version_info[0], 10, 0)
	minor, _ := strconv.ParseInt(version_info[1], 10, 0)
	return major, minor, nil
}
