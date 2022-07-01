# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

import pointermap

opt = pointermap.GetOptions()
print("opt = %r" % (opt,))

optmap = pointermap.GetOptionsMap()
print("optmap['opt'] = %r" % (optmap['opt'],))

#import pdb;pdb.set_trace()
print("optmap = %r" % (optmap,))

p = pointermap.NewProject("myproject")
print("p = %r" % (p,))
print("p.Name = %s" % (p.Name,))
print("p.Envs = %s" % (p.Envs,))

print("OK")
