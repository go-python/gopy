# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import go, funcs

fs = funcs.FunStruct()
fs.FieldS = "str field"
fs.FieldI = 42

def cbfun(afs, ival, sval):
    tfs = funcs.FunStruct(handle=afs)
    print("in python cbfun: FieldI: ", tfs.FieldI, " FieldS: ", tfs.FieldS, " ival: ", ival, " sval: ", sval)

def cbfunif(afs, ival, ifval):
    tfs = funcs.FunStruct(handle=afs)
    print("in python cbfunif: FieldI: ", tfs.FieldI, " FieldS: ", tfs.FieldS, " ival: ", ival, " ifval: ", ifval)

def cbfunrval(afs, ival, ifval):
    tfs = funcs.FunStruct(handle=afs)
    print("in python cbfunrval: FieldI: ", tfs.FieldI, " FieldS: ", tfs.FieldS, " ival: ", ival, " ifval: ", ifval)
    return True
    
class MyClass(go.GoClass):
    def __init__(self, *args, **kwargs):
        self.misc = 2
    
    def ClassFun(self, afs, ival, sval):
        tfs = funcs.FunStruct(handle=afs)
        print("in python class fun: FieldI: ", tfs.FieldI, " FieldS: ", tfs.FieldS, " ival: ", ival, " sval: ", sval)

    def CallSelf(self):
        fs.CallBack(77, self.ClassFun)
    
print("fs.CallBack(22, cbfun)...")
fs.CallBack(22, cbfun)

print("fs.CallBackIf(22, cbfunif)...")
fs.CallBackIf(22, cbfunif)

print("fs.CallBackRval(22, cbfunrval)...")
fs.CallBackRval(22, cbfunrval)

cls = MyClass()

# note: no special code needed to work with methods in callback (PyObject_CallObject just works)
# BUT it does NOT work if the callback is initiated from a different thread!  Then only regular
# functions work.
print("fs.CallBack(32, cls.ClassFun)...")
fs.CallBack(32, cls.ClassFun)

print("cls.CallSelf...")
cls.CallSelf()


print("fs.ObjArg with nil")
fs.ObjArg(go.nil)

print("fs.ObjArg with fs")
fs.ObjArg(fs)

# TODO: not currently supported:

# print("funcs.F1()...")
# f1 = funcs.F1()
# print("f1()= %s" % f1())
# 
# print("funcs.F2()...")
# f2 = funcs.F2()
# print("f2()= %s" % f2())
# 
# print("s1 = funcs.S1()...")
# s1 = funcs.S1()
# print("s1.F1 = funcs.F2()...")
# s1.F1 = funcs.F2()
# print("s1.F1() = %s" % s1.F1())
# 
# print("s2 = funcs.S2()...")
# s2 = funcs.S2()
# print("s2.F1 = funcs.F1()...")
# s2.F1 = funcs.F1()
# print("s2.F1() = %s" % s2.F1())

print("OK")
