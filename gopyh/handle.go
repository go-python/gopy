// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gopyh provides the variable handle manager for gopy.
// The handles map can NOT be globally shared in C because it
// must use interface{} values that can change location via GC.
// In effect, each gopy package must be thought of as a completely
// separate Go instance, and there can be NO sharing of anything
// between them, because they fundamentally live in different .so
// libraries.
// Thus, we must ensure that all handles used within a package
// are registered within that same package -- this means python
// users typically will import a single package, which exports
// all the relevant functionality as needed.  Any further packages
// cannot share anything other than basic types (int, string etc).
package gopyh

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

// type for the handle -- int64 for speed (can switch to string)
type GoHandle int64
type CGoHandle int64

// --- variable handles: all pointers managed via handles ---

var (
	ctr     int64
	handles map[GoHandle]interface{}
	mu      sync.RWMutex
)

func IfaceIsNil(it interface{}) bool {
	if it == nil {
		return true
	}
	v := reflect.ValueOf(it)
	vk := v.Kind()
	if vk == reflect.Ptr || vk == reflect.Interface || vk == reflect.Map || vk == reflect.Slice || vk == reflect.Func || vk == reflect.Chan {
		return v.IsNil()
	}
	return false
}

// Register registers a new variable instance
func Register(typnm string, ifc interface{}) CGoHandle {
	if IfaceIsNil(ifc) {
		return -1
	}
	mu.Lock()
	defer mu.Unlock()
	if handles == nil {
		handles = make(map[GoHandle]interface{})
	}
	ctr++
	hc := ctr
	handles[GoHandle(hc)] = ifc
	// fmt.Printf("gopy Registered: %s %d\n", typnm, hc)
	return CGoHandle(hc)
}

// VarFmHandle gets variable from handle string.
// Reports error to python but does not return it,
// for use in inline calls
func VarFmHandle(h CGoHandle, typnm string) interface{} {
	v, _ := VarFmHandleTry(h, typnm)
	return v
}

// VarFmHandleTry version returns the error explicitly,
// for use when error can be processed
func VarFmHandleTry(h CGoHandle, typnm string) (interface{}, error) {
	if h < 1 {
		return nil, fmt.Errorf("gopy: nil handle")
	}
	mu.RLock()
	defer mu.RUnlock()
	v, has := handles[GoHandle(h)]
	if !has {
		err := fmt.Errorf("gopy: variable handle not registered: " + strconv.FormatInt(int64(h), 10))
		// todo: need to get access to this:
		// C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
		return nil, err
	}
	return v, nil
}

/*
// --- string version of handle functions -- less opaque
// need to switch type in gopy/gen.c ---
func Register(typnm string, ifc interface{}) CGoHandle {
	if IfaceIsNil(ifc) {
		return C.CString("nil")
	}
	mu.Lock()
	defer mu.Unlock()
	if handles == nil {
		handles = make(map[GoHandle]interface{})
	}
	ctr++
	hc := ctr
	rnm := typnm + "_" + strconv.FormatInt(hc, 16)
	handles[GoHandle(rnm)] = ifc
	return C.CString(rnm)
}

// Try version returns the error explicitly, for use when error can be processed
func VarFmHandleTry(h CGoHandle, typnm string) (interface{}, error) {
	hs := C.GoString(h)
	if hs == "" || hs == "nil" {
		return nil, fmt.Errorf("gopy: nil handle")
	}
	mu.RLock()
	defer mu.RUnlock()
	// if !strings.HasPrefix(hs, typnm) {
	//  	err := fmt.Errorf("gopy: variable handle is not the correct type, should be: " + typnm + " is: " +  hs)
	//  	C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
	//  	return nil, err
	// }
	v, has := handles[GoHandle(hs)]
	if !has {
		err := fmt.Errorf("gopy: variable handle not registered: " +  hs)
		C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
		return nil, err
	}
	return v, nil
}
*/
