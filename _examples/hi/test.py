# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function

import hi

# NOTE: the output from python2 is different from that of 3, and test
# targets python3 so it appears to fail for 2.

print("--- doc(hi)...")
print(hi.__doc__)

print("--- hi.Universe:", hi.Universe)
print("--- hi.Version:", hi.Version)

print("--- hi.Debug():",hi.Debug())
print("--- hi.Set_Debug(true)")
hi.Set_Debug(True)
print("--- hi.Debug():",hi.Debug())
print("--- hi.Set_Debug(false)")
hi.Set_Debug(False)
print("--- hi.Debug():",hi.Debug())

print("--- hi.Anon():",hi.Anon())
anon = hi.NewPerson('you',24)
print("--- new anon:",anon)
print("--- hi.Set_Anon(hi.NewPerson('you', 24))...")
hi.Set_Anon(anon)
print("--- hi.Anon():",hi.Anon())

print("--- doc(hi.Hi)...")
print(hi.Hi.__doc__)

print("--- hi.Hi()...")
hi.Hi()

print("--- doc(hi.Hello)...")
print(hi.Hello.__doc__)

print("--- hi.Hello('you')...")
hi.Hello("you")

print("--- doc(hi.Add)...")
print(hi.Add.__doc__)

print("--- hi.Add(1, 41)...")
print(hi.Add(1,41))

print("--- hi.Concat('4', '2')...")
print(hi.Concat("4","2"))

print("--- hi.LookupQuestion(42)...")
print(hi.LookupQuestion(42))

print("--- hi.LookupQuestion(12)...")
try:
    hi.LookupQuestion(12)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err)
    pass

print("--- doc(hi.Person):")
print(hi.Person.__doc__)

print("--- p = hi.Person()...")
p = hi.Person()
print("--- p:", p)

print("--- p.Name:", p.Name)
print("--- p.Age:",p.Age)

print("--- doc(hi.Greet):")
print(p.Greet.__doc__)
print("--- p.Greet()...")
print(p.Greet())

print("--- p.String()...")
print(p.String())

print("--- doc(p):")
print(p.__doc__)

print("--- p.Name = \"foo\"...")
p.Name = "foo"

print("--- p.Age = 42...")
p.Age = 42

print("--- p.String()...")
print(p.String())
print("--- p.Age:", p.Age)
print("--- p.Name:",p.Name)

print("--- p.Work(2)...")
p.Work(2)

print("--- p.Work(24)...")
try:
    p.Work(24)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err)
    pass

print("--- p.Salary(2):", p.Salary(2))
try:
    print("--- p.Salary(24):",p.Salary(24))
    print("*ERROR* no exception raised!")
except Exception as err:
    print("--- p.Salary(24): caught:", err)
    pass

## test ctor args
print("--- Person.__init__")
try:
    hi.Person(1)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

try:
    hi.Person("name","2")
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

try:
    hi.Person("name",2,3)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

p = hi.Person("name")
print(p)
p = hi.Person("name", 42)
print(p)
p = hi.Person(Name="name", Age=42)
print(p)
p = hi.Person(Age=42, Name="name")
print(p)

## test ctors
print("--- hi.NewPerson('me', 666):", hi.NewPerson("me", 666))
print("--- hi.NewPersonWithAge(666):", hi.NewPersonWithAge(666))
print("--- hi.NewActivePerson(4):")
p = hi.NewActivePerson(4)
print(p)

## test Couple
print("--- c = hi.Couple()...")
c = hi.Couple()
print(c)
print("--- c.P1:", c.P1)
c.P1 = hi.NewPerson("tom", 5)
c.P2 = hi.NewPerson("bob", 2)
print("--- c:", c)

print("--- c = hi.NewCouple(tom, bob)...")
c = hi.NewCouple(hi.NewPerson("tom", 50), hi.NewPerson("bob", 41))
print(c)
c.P1.Name = "mom"
c.P2.Age = 51
print(c)

## test Couple.__init__
print("--- Couple.__init__")
# Note: pybindgen does not automatically support varargs, so in general
# all python calls need to provide the full Go signature of args.
#c = hi.Couple(hi.Person("p1", 42))
#print(c)
c = hi.Couple(hi.Person("p1", 42), hi.Person("p2", 52))
print(c)
c = hi.Couple(P1=hi.Person("p1", 42), P2=hi.Person("p2", 52))
print(c)
c = hi.Couple(P2=hi.Person("p1", 42), P1=hi.Person("p2", 52))
print(c)

try:
    hi.Couple(1)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

try:
    hi.Couple(1, 2)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

try:
    hi.Couple(P2=1)
    print("*ERROR* no exception raised!")
except Exception as err:
    print("caught:", err, "| err-type:",type(err))
    pass

### test gc
print("--- testing GC...")
NMAX = 100000
objs = []
for i in range(NMAX):
    p1  = hi.NewPerson("p1-%d" % i, i)
    p2 = hi.NewPerson("p2-%d" % i, i)
    objs.append(hi.NewCouple(p1,p2))
    pass
print("--- len(objs):",len(objs))
vs = []
for i,o in enumerate(objs):
    v = "%d: %s" % (i, o)
    vs.append(v)
    pass
print("--- len(vs):",len(vs))
del objs
print("--- testing GC... [ok]")

print("--- testing array...")
arr = hi.IntArray()
print("arr:",arr)
print("len(arr):",len(arr))
print("arr[0]:",arr[0])
print("arr[1]:",arr[1])
try:
    print("arr[2]:", arr[2])
    print("*ERROR* no exception raised!")
except Exception as err:
    print("arr[2]: caught:",err)
    pass
arr[1] = 42
print("arr:",arr)
print("len(arr):",len(arr))
try:
    print("mem(arr):",len(memoryview(arr)))
except Exception as err:
    print("mem(arr): caught:",err)
    pass

print("--- testing slice...")
s = hi.IntSlice()
print("slice:",s)
print("len(slice):",len(s))
print("slice[0]:",s[0])
print("slice[1]:",s[1])
try:
    print("slice[2]:", s[2])
    print("*ERROR* no exception raised!")
except Exception as err:
    print("slice[2]: caught:",err)
    pass
s[1] = 42
print("slice:",s)
print("slice repr:",s.__repr__())
print("len(slice):",len(s))
try:
    print("mem(slice):",len(memoryview(s)))
except Exception as err:
    print("mem(slice): caught:",err)
    pass

print("OK")
