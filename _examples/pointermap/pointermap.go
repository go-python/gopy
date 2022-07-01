// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pointermap

type Options map[string]interface{}

type Profile struct {
	Name       string
	Envs       map[string]Options
}

type Project struct {
	Profile *Profile
}

func NewProject(name string) Project{
	profile := Profile {
		Name: name,
		Envs: make(map[string]Options),
	}
	options := make(map[string]interface{})
	options["dev"] = 4
	profile.Envs["myPrjProfile"] = options
	prj := Project {
		Profile: &profile,
	}

	return prj
}

func GetOptions() map[string]interface{} {
	retval := make(map[string]interface{})
	retval["a"] = 1
	retval["b"] = "2"
	retval["c"] = true
	return retval
}

func GetOptionsMap() map[string]Options {
	retval := make(map[string]Options)
	retval["opt"] = GetOptions()

	return retval
}
