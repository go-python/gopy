# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

from funcs import go, funcs

def cbfun():
    print("in python cbfun!!")

class MyClass(object):
    def __init__(self, *args, **kwargs):
        self.misc = 2
    
    def ClassFun(self):
        print("in python class fun!!")
        
    
fs = funcs.FunStruct()
fs.CallBack(22, cbfun)

cls = MyClass()

fs.CallBack(22, cls.ClassFun)

# print("funcs.GetF1()...")
# f1 = funcs.GetF1()
# print("f1()= %s" % f1())
# 
# print("funcs.GetF2()...")
# f2 = funcs.GetF2()
# print("f2()= %s" % f2())
# 
# print("s1 = funcs.S1()...")
# s1 = funcs.S1()
# print("s1.F1 = funcs.GetF2()...")
# s1.F1 = funcs.GetF2()
# print("s1.F1() = %s" % s1.F1())
# 
# print("s2 = funcs.S2()...")
# s2 = funcs.S2()
# print("s2.F1 = funcs.GetF1()...")
# s2.F1 = funcs.GetF1()
# print("s2.F1() = %s" % s2.F1())
# 
#print("OK")
