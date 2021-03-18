// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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

type PyConfig struct {
	Version   int
	CFlags    string
	LdFlags   string
	ExtSuffix string
}

// AllFlags returns CFlags + " " + LdFlags
func (pc *PyConfig) AllFlags() string {
	return strings.TrimSpace(pc.CFlags) + " " + strings.TrimSpace(pc.LdFlags)
}

// GetPythonConfig returns the needed python configuration for the given
// python VM (python, python2, python3, pypy, etc...)
func GetPythonConfig(vm string) (PyConfig, error) {
	code := `import sys
import distutils.sysconfig as ds
import json
import os
version=sys.version_info.major

if "GOPY_INCLUDE" in os.environ and "GOPY_LIBDIR" in os.environ and "GOPY_PYLIB" in os.environ:
	print(json.dumps({
		"version": version,
		"incdir": os.environ["GOPY_INCLUDE"],
		"libdir": os.environ["GOPY_LIBDIR"],
		"libpy": os.environ["GOPY_PYLIB"],
		"shlibs":  ds.get_config_var("SHLIBS"),
		"syslibs": ds.get_config_var("SYSLIBS"),
		"shlinks": ds.get_config_var("LINKFORSHARED"),
		"extsuffix": ds.get_config_var("EXT_SUFFIX"),
}))
else:
	print(json.dumps({
		"version": sys.version_info.major,
		"incdir":  ds.get_python_inc(),
		"libdir":  ds.get_config_var("LIBDIR"),
		"libpy":   ds.get_config_var("LIBRARY"),
		"shlibs":  ds.get_config_var("SHLIBS"),
		"syslibs": ds.get_config_var("SYSLIBS"),
		"shlinks": ds.get_config_var("LINKFORSHARED"),
		"extsuffix": ds.get_config_var("EXT_SUFFIX"),
}))
`

	var cfg PyConfig
	bin, err := exec.LookPath(vm)
	if err != nil {
		return cfg, errors.Wrapf(err, "could not locate python vm %q", vm)
	}

	buf := new(bytes.Buffer)
	cmd := exec.Command(bin, "-c", code)
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return cfg, errors.Wrap(err, "could not run python-config script")
	}

	var raw struct {
		Version   int    `json:"version"`
		IncDir    string `json:"incdir"`
		LibDir    string `json:"libdir"`
		LibPy     string `json:"libpy"`
		ShLibs    string `json:"shlibs"`
		SysLibs   string `json:"syslibs"`
		ExtSuffix string `json:"extsuffix"`
	}
	err = json.NewDecoder(buf).Decode(&raw)
	if err != nil {
		return cfg, errors.Wrapf(err, "could not decode JSON script output")
	}

	raw.IncDir = filepath.ToSlash(raw.IncDir)
	raw.LibDir = filepath.ToSlash(raw.LibDir)

	if strings.HasSuffix(raw.LibPy, ".a") {
		raw.LibPy = raw.LibPy[:len(raw.LibPy)-len(".a")]
	}
	if strings.HasPrefix(raw.LibPy, "lib") {
		raw.LibPy = raw.LibPy[len("lib"):]
	}

	cfg.Version = raw.Version
	cfg.ExtSuffix = raw.ExtSuffix
	cfg.CFlags = strings.Join([]string{
		"-I" + raw.IncDir,
	}, " ")
	cfg.LdFlags = strings.Join([]string{
		"-L" + raw.LibDir,
		"-l" + raw.LibPy,
		raw.ShLibs,
		raw.SysLibs,
	}, " ")

	return cfg, nil
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

func extractPythonName(gname, gdoc string) (string, string, error) {
	const PythonName = "\ngopy:name "
	i := strings.Index(gdoc, PythonName)
	if i < 0 {
		return gname, gdoc, nil
	}
	s := gdoc[i+len(PythonName):]
	if end := strings.Index(s, "\n"); end > 0 {
		validIdPattern := regexp.MustCompile(`^[\pL_][\pL_\pN]+$`)
		if !validIdPattern.MatchString(s[:end]) {
			return "", "", fmt.Errorf("gopy: invalid identifier: %s", s[:end])
		}
		return s[:end], gdoc[:i] + s[end:], nil
	}
	return gname, gdoc, nil
}
