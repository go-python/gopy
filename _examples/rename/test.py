# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import rename

print("say_hi_fn():", rename.say_hi_fn())
print("MyStruct().say_something():", rename.MyStruct().say_something())

# Just make sure the symbols exist
rename.auto_renamed_func()
struct = rename.MyStruct()
struct.auto_renamed_meth()
_ = struct.auto_renamed_property
struct.auto_renamed_property = "foo"
_ = struct.custom_name
struct.custom_name = "foo"

print("MyStruct.auto_renamed_property.__doc__:", rename.MyStruct.auto_renamed_property.__doc__.strip())
print("MyStruct.custom_name.__doc__:", rename.MyStruct.custom_name.__doc__.strip())

print("OK")
