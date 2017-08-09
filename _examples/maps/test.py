# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import maps

a = maps.New()
b = {1: 3.0, 2: 5.0}
print('maps.Sum from Go map:', maps.Sum(a))
print('maps.Sum from Python dictionary:', maps.Sum(b))
print('maps.Keys from Go map:', maps.Keys(a))
print('maps.Values from Go map:', maps.Values(a))
print('maps.Keys from Python dictionary:', maps.Keys(b))
print('maps.Values from Python dictionary:', maps.Values(b))
