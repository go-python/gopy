package goerr2pyex

//#include <Python.h>
import "C"
import (
	"unsafe"
)

func DefaultGoErrorToPyException(err error) bool {
	if err != nil {
		estr := C.CString(err.Error())
		C.PyErr_SetString(C.PyExc_RuntimeError, estr)
		C.free(unsafe.Pointer(estr))
		return true
	} else {
		return false
	}
}
