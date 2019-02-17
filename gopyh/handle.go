// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gopyh provides the global variable handle manager
// for gopy.  this must be global to enable objects to be shared
// across different packages.
package gopyh

import "C"

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

type VarHandler struct {
	ctr  int64
	vars map[GoHandle]interface{}
	m    sync.RWMutex
}

// VarHand is the global variable handler instance
var VarHand VarHandler

func init() {
	if VarHand.ctr == 0 {
		VarHand.ctr = 1
	}
}

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

// VarFmHandle gets variable from handle string.
// Reports error to python but does not return it,
// for use in inline calls
func (vh *VarHandler) VarFmHandle(h CGoHandle, typnm string) interface{} {
	v, _ := vh.VarFmHandleTry(h, typnm)
	return v
}

// --- int64 version of handle functions ---

// Register registers a new variable instance
func (vh *VarHandler) Register(typnm string, ifc interface{}) CGoHandle {
	if IfaceIsNil(ifc) {
		return -1
	}
	vh.m.Lock()
	defer vh.m.Unlock()
	if vh.vars == nil {
		vh.vars = make(map[GoHandle]interface{})
	}
	hc := vh.ctr
	vh.ctr++
	vh.vars[GoHandle(hc)] = ifc
	fmt.Printf("gopy Registered: %s %d\n", typnm, hc)
	return CGoHandle(hc)
}

// VarFmHandleTry version returns the error explicitly,
// for use when error can be processed
func (vh *VarHandler) VarFmHandleTry(h CGoHandle, typnm string) (interface{}, error) {
	if h < 1 {
		return nil, fmt.Errorf("gopy: nil handle")
	}
	vh.m.RLock()
	defer vh.m.RUnlock()
	v, has := vh.vars[GoHandle(h)]
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

func (vh *VarHandler) Register(typnm string, ifc interface{}) CGoHandle {
	if IfaceIsNil(ifc) {
		return C.CString("nil")
	}
	vh.m.Lock()
	defer vh.m.Unlock()
	if vh.vars == nil {
		vh.vars = make(map[GoHandle]interface{})
	}
	hc := vh.ctr
	vh.ctr++
	rnm := typnm + "_" + strconv.FormatInt(hc, 16)
	vh.vars[GoHandle(rnm)] = ifc
	return C.CString(rnm)
}

// Try version returns the error explicitly, for use when error can be processed
func (vh *VarHandler) VarFmHandleTry(h CGoHandle, typnm string) (interface{}, error) {
	hs := C.GoString(h)
	if hs == "" || hs == "nil" {
		return nil, fmt.Errorf("gopy: nil handle")
	}
	vh.m.RLock()
	defer vh.m.RUnlock()
	// if !strings.HasPrefix(hs, typnm) {
	//  	err := fmt.Errorf("gopy: variable handle is not the correct type, should be: " + typnm + " is: " +  hs)
	//  	C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
	//  	return nil, err
	// }
	v, has := vh.vars[GoHandle(hs)]
	if !has {
		err := fmt.Errorf("gopy: variable handle not registered: " +  hs)
		C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
		return nil, err
	}
	return v, nil
}
*/
