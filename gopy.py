# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

__doc__ = """gopy is a convenience module to wrap and bind a Go package"""
__author__ = "The go-python authors"

__all__ = [
        "load",
        ]

### stdlib imports ---
import imp
import os
import sys

def printf(args):
    sys.stdout.write(args)
    pass

def load(pkg, output=""):
    """
    `load` takes a fully qualified Go package name and runs `gopy bind` on it. 
    @returns the C-extension module object
    """
    printf("::: gopy.loading '%s'...\n" % pkg)
    
    from subprocess import check_call
    if output == "": 
        output = os.getcwd()
        pass

    check_call(["gopy","bind", "-output=%s" % output, pkg])
    
    n = os.path.basename(pkg)
    printf("::: gopy.module=gopy.%s\n" % n)
    
    ok = imp.find_module(n, [output])
    if not ok:
        raise RuntimeError("could not find module 'gopy.%s'" % n)
    fname, path, descr = ok
    return imp.load_module("gopy."+n, fname, path, descr)
    

    

