// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cpy implements the seq serialization protocol as defined by
// golang.org/x/mobile/bind/seq.
//
// See the design document (http://golang.org/s/gobind).
package main

//#include <stdint.h>
//#include <stddef.h>
//#include <stdlib.h>
//#include <string.h>
import "C"

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/mobile/bind/seq"
)

const maxSliceLen = 1<<31 - 1

const debug = false

// cgopy_seq_send is called by CPython to send a request to run a Go function.
//export cgopy_seq_send
func cgopy_seq_send(descriptor *C.char, code int, req *C.uint8_t, reqlen C.uint32_t, res **C.uint8_t, reslen *C.uint32_t) {
	descr := C.GoString(descriptor)
	fmt.Fprintf(os.Stderr, "descr=%q, code=%d, req=%p, len=%d...\n", descr, code, req, reqlen)
	fn := seq.Registry[descr][code]
	if fn == nil {
		panic(fmt.Sprintf("gopy: invalid descriptor(%s) and code(0x%x)", descr, code))
	}
	in := new(seq.Buffer)
	if reqlen > 0 {
		in.Data = (*[maxSliceLen]byte)(unsafe.Pointer(req))[:reqlen]
	}
	out := new(seq.Buffer)
	fn(out, in)
	// BUG(hyangah): the function returning a go byte slice (so fn writes a pointer into 'out') is unsafe.
	// After fn is complete here, Go runtime is free to collect or move the pointed byte slice
	// contents. (Explicitly calling runtime.GC here will surface the problem?)
	// Without pinning support from Go side, it will be hard to fix it without extra copying.

	seqToBuf(res, reslen, out)
}

// cgopy_seq_destroy_ref is called by CPython to inform Go it is done with a reference.
//export cgopy_seq_destroy_ref
func cgopy_seq_destroy_ref(refnum C.int32_t) {
	seq.Delete(int32(refnum))
}

type request struct {
	ref    *seq.Ref
	handle int32
	code   int
	in     *seq.Buffer
}

var recv struct {
	sync.Mutex
	cond sync.Cond // signals req is not empty
	req  []request
	next int32 // next handle value
}

var res struct {
	sync.Mutex
	cond sync.Cond             // signals a response is filled in
	out  map[int32]*seq.Buffer // handle -> output
}

func init() {
	recv.cond.L = &recv.Mutex
	recv.next = 411 // arbitrary starting point distinct from Go and CPython obj ref nums

	res.cond.L = &res.Mutex
	res.out = make(map[int32]*seq.Buffer)
}

func seqToBuf(bufptr **C.uint8_t, lenptr *C.uint32_t, buf *seq.Buffer) {
	if debug {
		fmt.Printf("gopy: seqToBuf tag 1, len(buf.Data)=%d, *lenptr=%d\n", len(buf.Data), *lenptr)
	}
	if len(buf.Data) == 0 {
		*lenptr = 0
		return
	}
	if len(buf.Data) > int(*lenptr) {
		// TODO(crawshaw): realloc
		C.free(unsafe.Pointer(*bufptr))
		m := C.malloc(C.size_t(len(buf.Data)))
		if uintptr(m) == 0 {
			panic(fmt.Sprintf("gopy: malloc failed, size=%d", len(buf.Data)))
		}
		*bufptr = (*C.uint8_t)(m)
		*lenptr = C.uint32_t(len(buf.Data))
	}
	C.memcpy(unsafe.Pointer(*bufptr), unsafe.Pointer(&buf.Data[0]), C.size_t(len(buf.Data)))
}

// transact calls a method on a CPython object instance.
// It blocks until the call is complete.
func transact(ref *seq.Ref, _ string, code int, in *seq.Buffer) *seq.Buffer {
	recv.Lock()
	if recv.next == 1<<31-1 {
		panic("gopy: recv handle overflow")
	}
	handle := recv.next
	recv.next++
	recv.req = append(recv.req, request{
		ref:    ref,
		code:   code,
		in:     in,
		handle: handle,
	})
	recv.Unlock()
	recv.cond.Signal()

	res.Lock()
	for res.out[handle] == nil {
		res.cond.Wait()
	}
	out := res.out[handle]
	delete(res.out, handle)
	res.Unlock()

	return out
}

func encodeString(out *seq.Buffer, v string) {
	// FIXME(sbinet): cpython3 uses utf-8. cpython2 does not.
	//out.WriteUTF8(v)
	out.WriteByteArray([]byte(v))
}

func decodeString(in *seq.Buffer) string {
	// FIXME(sbinet): cpython3 uses utf-8. cpython2 does not.
	//return in.ReadUTF8()
	return string(in.ReadByteArray())
}

func init() {
	seq.FinalizeRef = func(ref *seq.Ref) {
		if ref.Num < 0 {
			panic(fmt.Sprintf("gopy: not a CPython ref: %d", ref.Num))
		}
		transact(ref, "", -1, new(seq.Buffer))
	}

	seq.Transact = transact
	seq.EncString = encodeString
	seq.DecString = decodeString
}
