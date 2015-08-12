# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

### py2/py3 compat
from __future__ import print_function

__doc__ = """gopy is a convenience module to wrap and bind a Go package"""
__author__ = "The go-python authors"

__all__ = [
        "load",
        ]

### stdlib imports ---
import imp
import os
import sys

def load(pkg, output=""):
    """
    `load` takes a fully qualified Go package name and runs `gopy bind` on it. 
    @returns the C-extension module object
    """

    from subprocess import check_call, check_output
    if output == "": 
        output = os.getcwd()
        pass

    print("gopy> inferring package name...")
    pkg = check_output(["go", "list", pkg]).strip()
    if pkg in sys.modules:
        print("gopy> package '%s' already wrapped and loaded!" % (pkg,))
        print("gopy> NOT recompiling it again (see issue #27)")
        return sys.modules[pkg]
    print("gopy> loading '%s'..." % pkg)

    check_call(["gopy","bind", "-output=%s" % output, pkg])
    
    n = os.path.basename(pkg)
    print("gopy> importing '%s'" % (pkg,))
    
    ok = imp.find_module(n, [output])
    if not ok:
        raise RuntimeError("could not find module '%s'" % pkg)
    fname, path, descr = ok
    mod = imp.load_module('__gopy__.'+n, fname, path, descr)
    mod.__name__ = pkg
    sys.modules[pkg] = mod
    del sys.modules['__gopy__.'+n]
    return mod
    

    

