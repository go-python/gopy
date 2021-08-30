// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/go-python/gopy/bind"
)

var (
	testBackends = map[string]string{}
	features     = map[string][]string{
		"_examples/hi":          []string{"py3"}, // output is different for 2 vs. 3 -- only checking 3 output
		"_examples/funcs":       []string{"py2", "py3"},
		"_examples/sliceptr":    []string{"py2", "py3"},
		"_examples/simple":      []string{"py2", "py3"},
		"_examples/empty":       []string{"py2", "py3"},
		"_examples/named":       []string{"py2", "py3"},
		"_examples/structs":     []string{"py2", "py3"},
		"_examples/consts":      []string{"py2", "py3"}, // 2 doesn't report .666 decimals
		"_examples/vars":        []string{"py2", "py3"},
		"_examples/seqs":        []string{"py2", "py3"},
		"_examples/cgo":         []string{"py2", "py3"},
		"_examples/pyerrors":    []string{"py2", "py3"},
		"_examples/iface":       []string{"py3"}, // output order diff for 2, fails but actually works
		"_examples/pointers":    []string{"py2", "py3"},
		"_examples/arrays":      []string{"py2", "py3"},
		"_examples/slices":      []string{"py2", "py3"},
		"_examples/maps":        []string{"py2", "py3"},
		"_examples/gostrings":   []string{"py2", "py3"},
		"_examples/rename":      []string{"py2", "py3"},
		"_examples/lot":         []string{"py2", "py3"},
		"_examples/unicode":     []string{"py3"}, // doesn't work for 2
		"_examples/osfile":      []string{"py2", "py3"},
		"_examples/gopygc":      []string{"py2", "py3"},
		"_examples/cstrings":    []string{"py2", "py3"},
		"_examples/pkgconflict": []string{"py2", "py3"},
	}

	testEnvironment = os.Environ()
)

func init() {
	os.Setenv("GOFLAGS", "-mod=mod")
	testEnvironment = append(testEnvironment, "GOFLAGS=-mod=mod")
}

func TestGovet(t *testing.T) {
	cmd := exec.Command("go", "vet", "./...")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("error running %s:\n%s\n%v", "go vet", string(buf.Bytes()), err)
	}
}

func TestGofmt(t *testing.T) {
	exe, err := exec.LookPath("goimports")
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			if e.Err == exec.ErrNotFound {
				exe, err = exec.LookPath("gofmt")
			}
		}
	}
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(exe, "-d", ".")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = buf

	err = cmd.Run()
	if err != nil {
		t.Fatalf("error running %s:\n%s\n%v", exe, string(buf.Bytes()), err)
	}

	if len(buf.Bytes()) != 0 {
		t.Errorf("some files were not gofmt'ed:\n%s\n", string(buf.Bytes()))
	}
}

func TestGoPyErrors(t *testing.T) {
	pyvm := testBackends["py3"]
	workdir, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		t.Fatalf("could not create workdir: %v\n", err)
	}
	t.Logf("pyvm: %s making work dir: %s\n", pyvm, workdir)
	defer os.RemoveAll(workdir)

	curPkgPath := reflect.TypeOf(pkg{}).PkgPath()
	fpath := filepath.Join(curPkgPath, "_examples/gopyerrors")
	cmd := exec.Command("go", "run", ".", "gen", "-vm="+pyvm, "-output="+workdir, fpath)
	t.Logf("running: %v\n", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not run %v: %+v\n", strings.Join(cmd.Args, " "), err)
	}
	contains := `--- Processing package: github.com/go-python/gopy/_examples/gopyerrors ---
ignoring python incompatible function: .func github.com/go-python/gopy/_examples/gopyerrors.NotErrorMany() (int, int): func() (int, int): gopy: second result value must be of type error: func() (int, int)
ignoring python incompatible method: gopyerrors.func (*github.com/go-python/gopy/_examples/gopyerrors.Struct).NotErrorMany() (int, string): func() (int, string): gopy: second result value must be of type error: func() (int, string)
ignoring python incompatible method: gopyerrors.func (*github.com/go-python/gopy/_examples/gopyerrors.Struct).TooMany() (int, int, string): func() (int, int, string): gopy: too many results to return: func() (int, int, string)
ignoring python incompatible function: .func github.com/go-python/gopy/_examples/gopyerrors.TooMany() (int, int, string): func() (int, int, string): gopy: too many results to return: func() (int, int, string)
`
	if got, want := string(out), contains; !strings.Contains(got, want) {
		t.Fatalf("%v does not contain\n%v\n", got, want)
	}
}

func TestHi(t *testing.T) {
	// t.Parallel()
	path := "_examples/hi"
	// NOTE: output differs for python2 -- only valid checking for 3
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`hi from go
hello you from go
working...
worked for 2 hours
working...
working...
worked for 4 hours
--- doc(hi)...

package hi exposes a few Go functions to be wrapped and used from Python.


--- hi.Universe: 42
--- hi.Version: 0.1
--- hi.Debug(): False
--- hi.Set_Debug(true)
--- hi.Debug(): True
--- hi.Set_Debug(false)
--- hi.Debug(): False
--- hi.Anon(): hi.Person{Name="<nobody>", Age=1}
--- new anon: hi.Person{Name="you", Age=24}
--- hi.Set_Anon(hi.NewPerson('you', 24))...
--- hi.Anon(): hi.Person{Name="you", Age=24}
--- doc(hi.Hi)...
Hi() 
	
	Hi prints hi from Go
	
--- hi.Hi()...
--- doc(hi.Hello)...
Hello(str s) 
	
	Hello prints a greeting from Go
	
--- hi.Hello('you')...
--- doc(hi.Add)...
Add(int i, int j) int
	
	Add returns the sum of its arguments.
	
--- hi.Add(1, 41)...
42
--- hi.Concat('4', '2')...
42
--- hi.LookupQuestion(42)...
Life, the Universe and Everything
--- hi.LookupQuestion(12)...
caught: Wrong answer: 12 != 42
--- doc(hi.Person):
Person is a simple struct

--- p = hi.Person()...
--- p: hi.Person{Name="", Age=0}
--- p.Name: 
--- p.Age: 0
--- doc(hi.Greet):
Greet() str
		
		Greet sends greetings
		
--- p.Greet()...
Hello, I am 
--- p.String()...
hi.Person{Name="", Age=0}
--- doc(p):
Person is a simple struct

--- p.Name = "foo"...
--- p.Age = 42...
--- p.String()...
hi.Person{Name="foo", Age=42}
--- p.Age: 42
--- p.Name: foo
--- p.Work(2)...
--- p.Work(24)...
caught: can't work for 24 hours!
--- p.Salary(2): 20
--- p.Salary(24): caught: can't work for 24 hours!
--- Person.__init__
caught: argument 2 must be str, not int | err-type: <class 'TypeError'>
caught: an integer is required (got type str) | err-type: <class 'TypeError'>
*ERROR* no exception raised!
hi.Person{Name="name", Age=0}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
--- hi.NewPerson('me', 666): hi.Person{Name="me", Age=666}
--- hi.NewPersonWithAge(666): hi.Person{Name="stranger", Age=666}
--- hi.NewActivePerson(4):
hi.Person{Name="", Age=0}
--- c = hi.Couple()...
hi.Couple{P1=hi.Person{Name="", Age=0}, P2=hi.Person{Name="", Age=0}}
--- c.P1: hi.Person{Name="", Age=0}
--- c: hi.Couple{P1=hi.Person{Name="tom", Age=5}, P2=hi.Person{Name="bob", Age=2}}
--- c = hi.NewCouple(tom, bob)...
hi.Couple{P1=hi.Person{Name="tom", Age=50}, P2=hi.Person{Name="bob", Age=41}}
hi.Couple{P1=hi.Person{Name="mom", Age=50}, P2=hi.Person{Name="bob", Age=51}}
--- Couple.__init__
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p2", Age=52}, P2=hi.Person{Name="p1", Age=42}}
caught: supplied argument type <class 'int'> is not a go.GoClass | err-type: <class 'TypeError'>
caught: supplied argument type <class 'int'> is not a go.GoClass | err-type: <class 'TypeError'>
caught: supplied argument type <class 'int'> is not a go.GoClass | err-type: <class 'TypeError'>
--- testing GC...
--- len(objs): 100000
--- len(vs): 100000
--- testing GC... [ok]
--- testing array...
arr: hi.Array_2_int len: 2 handle: 300036 [1, 2]
len(arr): 2
arr[0]: 1
arr[1]: 2
arr[2]: caught: slice index out of range
arr: hi.Array_2_int len: 2 handle: 300036 [1, 2]
len(arr): 2
mem(arr): caught: memoryview: a bytes-like object is required, not 'Array_2_int'
--- testing slice...
slice: go.Slice_int len: 2 handle: 300037 [1, 2]
len(slice): 2
slice[0]: 1
slice[1]: 2
slice[2]: caught: slice index out of range
slice: go.Slice_int len: 2 handle: 300037 [1, 42]
slice repr: go.Slice_int([1, 42])
len(slice): 2
mem(slice): caught: memoryview: a bytes-like object is required, not 'Slice_int'
OK
`),
	})

}

func TestBindFuncs(t *testing.T) {
	// t.Parallel()
	path := "_examples/funcs"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`got return value: true
got nil
ofs FieldI: 42 FieldS: str field
fs.CallBack(22, cbfun)...
in python cbfun: FieldI:  42  FieldS:  str field  ival:  22  sval:  str field
fs.CallBackIf(22, cbfunif)...
in python cbfunif: FieldI:  42  FieldS:  str field  ival:  22  ifval:  str field
fs.CallBackRval(22, cbfunrval)...
in python cbfunrval: FieldI:  42  FieldS:  str field  ival:  22  ifval:  str field
fs.CallBack(32, cls.ClassFun)...
in python class fun: FieldI:  42  FieldS:  str field  ival:  32  sval:  str field
cls.CallSelf...
in python class fun: FieldI:  42  FieldS:  str field  ival:  77  sval:  str field
fs.ObjArg with nil
fs.ObjArg with fs
OK
`),
	})
}

func TestBindSimple(t *testing.T) {
	// t.Parallel()
	path := "_examples/simple"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`doc(pkg):
'\nsimple is a simple package.\n\n'
pkg.Func()...
fct = pkg.Func...
fct()...
pkg.Add(1,2)= 3
pkg.Bool(True)= True
pkg.Bool(False)= False
pkg.Comp64Add((3+4j), (2+5j)) = (5+9j)
pkg.Comp128Add((3+4j), (2+5j)) = (5+9j)
OK
`),
	})
}

func TestBindEmpty(t *testing.T) {
	// t.Parallel()
	path := "_examples/empty"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`doc(pkg):
'\nPackage empty does not expose anything.\nWe may want to wrap and import it just for its side-effects.\n\n'
OK
`),
	})
}

func TestBindPointers(t *testing.T) {
	// t.Parallel()
	path := "_examples/pointers"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`s = pointers.S(2)
s = pointers.S{Value=2, handle=1}
s.Value = 2
OK
`),
	})
}

func TestBindNamed(t *testing.T) {
	// t.Parallel()
	path := "_examples/named"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`doc(named): '\npackage named tests various aspects of named types.\n\n'
s = named.Slice()
s = named.named_Slice len: 0 handle: 1 []
s = named.Slice([1,2])
s = named.named_Slice len: 2 handle: 2 [1.0, 2.0]
s = named.Slice(range(10))
s = named.named_Slice len: 10 handle: 3 [0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0]
s = named.Slice(xrange(10))
s = named.named_Slice len: 10 handle: 4 [0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0]
OK
`),
	})
}

func TestBindStructs(t *testing.T) {
	// t.Parallel()
	path := "_examples/structs"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`s = structs.S()
s = structs.S{handle=1}
s.Init()
s.Upper('boo')= 'BOO'
s1 = structs.S1()
s1 = structs.S1{handle=2}
caught error: 'S1' object has no attribute 'private'
s2 = structs.S2()
s2 = structs.S2{Public=1, handle=5}
s2 = structs.S2{Public=42, handle=7}
s2.Public = 42
caught error: 'S2' object has no attribute 'private'
s2child = S2Child{S2: structs.S2{Public=42, handle=8, local=123}, local: 123}
s2child.Public = 42
s2child.local = 123
caught error: 'S2Child' object has no attribute 'private'
s3.X,Y = 3,4
OK
`),
	})
}

func TestBindConsts(t *testing.T) {
	// t.Parallel()
	path := "_examples/consts"
	// NOTE: python2 does not properly output .666 decimals for unknown reasons.
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`c1 = c1
c2 = 42
c3 = 666.666
c4 = c4
c5 = 42
c6 = 42
c7 = 666.666
k1 = 1
k2 = 2
OK
`),
	})
}

func TestBindVars(t *testing.T) {
	// t.Parallel()
	path := "_examples/vars"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`doc(vars):
None
doc(vars.V1()):
'\n\tV1 Gets Go Variable: vars.V1\n\t\n\t'
doc(vars.Set_V1()):
'\n\tSet_V1 Sets Go Variable: vars.V1\n\t\n\t'
Initial values
v1 = v1
v2 = 42
v3 = 666.666
v4 = c4
v5 = 42
v6 = 42
v7 = 666.666
k1 = 1
k2 = 2
New values
v1 = test1
v2 = 90
v3 = 1111.1111
v4 = test2
v5 = 50
v6 = 50
v7 = 1111.1111
k1 = 123
k2 = 456
vars.Doc() = 'A variable with some documentation'
doc of vars.Doc = '\n\tDoc Gets Go Variable: vars.Doc\n\tDoc is a top-level string with some documentation attached.\n\t\n\t'
doc of vars.Set_Doc = '\n\tSet_Doc Sets Go Variable: vars.Doc\n\tDoc is a top-level string with some documentation attached.\n\t\n\t'
OK
`),
	})
}

func TestBindSeqs(t *testing.T) {
	// t.Parallel()
	path := "_examples/seqs"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`doc(seqs): '\npackage seqs tests various aspects of sequence types.\n\n'
s = seqs.Slice()
s = seqs.seqs_Slice len: 0 handle: 1 []
s = seqs.Slice([1,2])
s = seqs.seqs_Slice len: 2 handle: 2 [1.0, 2.0]
s = seqs.Slice(range(10))
s = seqs.seqs_Slice len: 10 handle: 3 [0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0]
s = seqs.Slice(xrange(10))
s = seqs.seqs_Slice len: 10 handle: 4 [0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0]
s = seqs.Slice()
s = seqs.seqs_Slice len: 0 handle: 5 []
s += [1,2]
s = seqs.seqs_Slice len: 2 handle: 5 [1.0, 2.0]
s += [10,20]
s = seqs.seqs_Slice len: 4 handle: 5 [1.0, 2.0, 10.0, 20.0]
OK
`),
	})
}

func TestBindInterfaces(t *testing.T) {
	// t.Parallel()
	path := "_examples/iface"
	// NOTE: python2 outputs this in a different order, so test fails, but results are the same
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`t.F [CALLED]
iface.CallIface...
t.F [CALLED]
iface.CallIface... [DONE]
iface as string: test string
iface as string: 42
iface as handle: &{0 }
iface as handle: <nil>
doc(iface): '\npackage iface tests various aspects of interfaces.\n\n'
t = iface.T()
t.F()
iface.CallIface(t)
iface.IfaceString("test string"
iface.IfaceString(str(42))
iface.IfaceHandle(t)
iface.IfaceHandle(go.nil)
OK
`),
	})
}

func TestBindCgoPackage(t *testing.T) {
	// t.Parallel()
	path := "_examples/cgo"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`cgo.doc: '\nPackage cgo tests bindings of CGo-based packages.\n\n'
cgo.Hi()= 'hi from go\n'
cgo.Hello(you)= 'hello you from go\n'
OK
`),
	})
}

func TestPyErrors(t *testing.T) {
	// t.Parallel()
	path := "_examples/pyerrors"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`Divide by zero.
pyerrors.Div(5, 2) = 2
Empty string value.
pyerrors.NewMyString("hello") = "hello"
OK
`),
	})
}

func TestBuiltinArrays(t *testing.T) {
	// t.Parallel()
	path := "_examples/arrays"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`Python list: [1, 2, 3, 4]
Go array:  arrays.Array_4_int len: 4 handle: 1 [1, 2, 3, 4]
arrays.IntSum from Go array: 10
OK
`),
	})
}

func TestBuiltinSlices(t *testing.T) {
	// t.Parallel()
	path := "_examples/slices"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`0: S0
1: S1
2: S2
S0
S1
Python list: [1, 2, 3, 4]
Go slice:  go.Slice_int len: 4 handle: 1 [1, 2, 3, 4]
slices.IntSum from Python list: 10
slices.IntSum from Go slice: 10
unsigned slice elements: 1 2 3 4
signed slice elements: -1 -2 -3 -4
struct slice:  slices.Slice_Ptr_slices_S len: 3 handle: 11 [12, 13, 14]
struct slice[0]:  slices.S{Name=S0, handle=15}
struct slice[1]:  slices.S{Name=S1, handle=16}
struct slice[2].Name:  S2
OK
`),
	})
}

func TestBuiltinMaps(t *testing.T) {
	// t.Parallel()
	path := "_examples/maps"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`map a[1]: 3.0
map a[2]: 5.0
2 in map: True
3 in map: False
maps.Sum from Go map: 8.0
map b: {1: 3.0, 2: 5.0}
maps.Sum from Python dictionary: 8.0
maps.Keys from Go map: go.Slice_int len: 2 handle: 3 [1, 2]
maps.Values from Go map: go.Slice_float64 len: 2 handle: 4 [3.0, 5.0]
maps.Keys from Python dictionary: go.Slice_int len: 2 handle: 6 [1, 2]
maps.Values from Python dictionary: go.Slice_float64 len: 2 handle: 8 [3.0, 5.0]
deleted 1 from a: maps.Map_int_float64 len: 1 handle: 1 {2=5.0, }
OK
`),
	})
}

func TestBindStrings(t *testing.T) {
	// t.Parallel()
	path := "_examples/gostrings"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`S1 = S1
GetString() = MyString
OK
`),
	})
}

func TestBindRename(t *testing.T) {
	// t.Parallel()
	path := "_examples/rename"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: []string{"-rename"},
		want: []byte(`say_hi_fn(): hi
MyStruct().say_something(): something
MyStruct.auto_renamed_property.__doc__: I should be renamed to auto_renamed_property
		when generated with -rename flag
MyStruct.custom_name.__doc__: I should be renamed to custom_name with the custom option
OK
`),
	})
}

func TestLot(t *testing.T) {
	// t.Parallel()
	path := "_examples/lot"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`l.SomeString : some string
l.SomeInt : 1337
l.SomeFloat : 1337.1337
l.SomeBool : True
l.SomeListOfStrings: go.Slice_string len: 4 handle: 2 [some, list, of, strings]
l.SomeListOfInts: go.Slice_int64 len: 4 handle: 3 [6, 2, 9, 1]
l.SomeListOfFloats: go.Slice_float64 len: 4 handle: 4 [6.6, 2.2, 9.9, 1.1]
l.SomeListOfBools: go.Slice_bool len: 4 handle: 5 [True, False, True, False]
OK
`),
	})
}

func TestSlicePtr(t *testing.T) {
	// t.Parallel()
	path := "_examples/sliceptr"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`sliceptr.sliceptr_IntVector len: 3 handle: 1 [1, 2, 3]
sliceptr.sliceptr_IntVector len: 4 handle: 1 [1, 2, 3, 4]
sliceptr.sliceptr_StrVector len: 4 handle: 2 [1, 2, 3, 4]
OK
`),
	})
}

func TestUnicode(t *testing.T) {
	// t.Parallel()
	path := "_examples/unicode"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`encoding.HandleString(unicodestr) -> Python Unicode string 🐱
encoding.GetString() -> Go Unicode string 🐱
OK
`),
	})
}

func TestPYGC(t *testing.T) {
	// t.Parallel()
	path := "_examples/gopygc"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`0
3
0
3
5
6
7
8
5
3
2
1
0
1
0
1
1
1
0
OK
`),
	})
}

func TestCStrings(t *testing.T) {
	// t.Parallel()
	path := "_examples/cstrings"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`gofnString leaked:  False
gofnStruct leaked:  False
gofnNestedStruct leaked:  False
gofnSlice leaked:  False
gofnMap leaked:  False
OK
`),
	})
}

func TestPackagePrefix(t *testing.T) {
	// t.Parallel()
	path := "_examples/package/mypkg"
	testPkg(t, pkg{
		path:      path,
		lang:      features[path],
		cmd:       "build",
		outputdir: "mypkg",
		pkgprefix: ".",
		testdir:   ".",
		want: []byte(`mypkg.mypkg.SayHello()...
Hello
OK
`),
	})
}

func TestPkgConflict(t *testing.T) {
	// t.Parallel()
	path := "_examples/pkgconflict"
	testPkg(t, pkg{
		path:   path,
		lang:   features[path],
		cmd:    "build",
		extras: nil,
		want: []byte(`17 items are made of wool
42 items are made of hyperlinks
2 items are made of fuzz
`),
	})
}

// see notes in _examples/osfile/test.py for why this doesn't work..
// leaving here for now in case someone wants to follow-up and make it work..
//
// func TestOsFile(t *testing.T) {
// 	// t.Parallel()
// 	path := "_examples/osfile"
// 	testPkg(t, pkg{
// 		path:   path,
// 		lang:   features[path],
// 		cmd:    "pkg",
// 		extras: []string{"os"},
// 		want: []byte(`yeah..
// OK
// `),
// 	})
// }

// Generate / verify SUPPORT_MATRIX.md from features map.
func TestCheckSupportMatrix(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("# Support matrix\n")
	buf.WriteString(`
NOTE: File auto-generated by TestCheckSupportMatrix in main_test.go. Please
don't modify manually.

`)

	// Generate
	// - sorted list of features
	// - sorted list of backends
	// - a map of feature to available backends
	var featuresSorted []string
	var allBackendsSorted []string
	featureToBackendMap := make(map[string]map[string]bool)
	allBackends := make(map[string]bool)
	for feature, backends := range features {
		featuresSorted = append(featuresSorted, feature)
		featureToBackendMap[feature] = make(map[string]bool)
		for _, backend := range backends {
			featureToBackendMap[feature][backend] = true
			allBackends[backend] = true
		}
	}
	for backend, _ := range allBackends {
		allBackendsSorted = append(allBackendsSorted, backend)
	}
	sort.Strings(featuresSorted)
	sort.Strings(allBackendsSorted)

	// Write the table header and the line separating header and rows.
	fmt.Fprintf(&buf, "Feature |%s\n", strings.Join(allBackendsSorted, " | "))
	var tableDelimiters []string
	for i := 0; i <= len(allBackendsSorted); i++ {
		tableDelimiters = append(tableDelimiters, "---")
	}
	buf.WriteString(strings.Join(tableDelimiters, " | "))
	buf.WriteString("\n")

	// Write the actual rows of the support matrix.
	for _, feature := range featuresSorted {
		var cells []string
		cells = append(cells, feature)
		for _, backend := range allBackendsSorted {
			if featureToBackendMap[feature][backend] {
				cells = append(cells, "yes")
			} else {
				cells = append(cells, "no")
			}
		}
		buf.WriteString(strings.Join(cells, " | "))
		buf.WriteString("\n")
	}

	if os.Getenv("GOPY_GENERATE_SUPPORT_MATRIX") == "1" {
		err := ioutil.WriteFile("SUPPORT_MATRIX.md", buf.Bytes(), 0644)
		if err != nil {
			log.Fatalf("Unable to write SUPPORT_MATRIX.md")
		}
		return
	}

	src, err := ioutil.ReadFile("SUPPORT_MATRIX.md")
	if err != nil {
		log.Fatalf("Unable to read SUPPORT_MATRIX.md")
	}

	msg := `
This is a test case to verify the support matrix. This test is likely failing
because the map features has been updated and the
auto-generated file SUPPORT_MATRIX.md hasn't been updated. Please run 'go test'
with environment variable GOPY_GENERATE_SUPPORT_MATRIX=1 to regenerate
SUPPORT_MATRIX.md and commit the changes to SUPPORT_MATRIX.md onto git.
`
	if bytes.Compare(buf.Bytes(), src) != 0 {
		t.Fatalf(msg)
	}
}

type pkg struct {
	path      string
	lang      []string
	cmd       string
	outputdir string
	pkgprefix string
	testdir   string
	extras    []string
	want      []byte
}

func testPkg(t *testing.T, table pkg) {
	// backends := []string{"py2", "py3"}
	backends := []string{"py3"}
	// backends := table.lang // todo: enabling py2 testing requires separate "want" output
	for _, be := range backends {
		fmt.Printf("looping over backends: %s in %s\n", be, backends)
		vm, ok := testBackends[be]
		if !ok || vm == "" {
			// backend not available.
			t.Logf("Skipped testing backend %s for %s\n", be, table.path)
			continue
		}
		switch be {
		case "py2":
			t.Run(be, func(t *testing.T) {
				// t.Parallel()
				testPkgBackend(t, vm, table)
			})
		case "py3":
			t.Run(be, func(t *testing.T) {
				// t.Parallel()
				testPkgBackend(t, vm, table)
			})
		default:
			t.Errorf("invalid backend name %q", be)
		}
	}
}

func writeGoMod(t *testing.T, pkgDir, tstDir string) {
	template := `
module dummy

require github.com/go-python/gopy v0.0.0
replace github.com/go-python/gopy => %s
`
	contents := fmt.Sprintf(template, pkgDir)
	if err := ioutil.WriteFile(filepath.Join(tstDir, "go.mod"), []byte(contents), 0666); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}
}

func testPkgBackend(t *testing.T, pyvm string, table pkg) {
	curPkgPath := reflect.TypeOf(table).PkgPath()
	_, pkgNm := filepath.Split(table.path)
	cwd, _ := os.Getwd()
	workdir, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		t.Fatalf("[%s:%s]: could not create workdir: %v\n", pyvm, table.path, err)
	}
	genPkgDir := workdir
	if table.outputdir != "" {
		genPkgDir = filepath.Join(workdir, table.outputdir)
		fmt.Printf("pyvm: %s making temp output dir: %s\n", pyvm, genPkgDir)
		err = os.MkdirAll(genPkgDir, 0700)
		if err != nil {
			t.Fatalf("[%s:%s]: could not create temp output dir: %v\n", pyvm, table.path, err)
		}
	}
	defer os.RemoveAll(workdir)
	defer bind.ResetPackages()

	env := make([]string, len(testEnvironment))
	copy(env, testEnvironment)
	env = append(env, fmt.Sprintf("PYTHONPATH=%s", workdir))

	writeGoMod(t, cwd, genPkgDir)

	// fmt.Printf("building in work dir: %s\n", workdir)
	fpath := "./" + table.path
	if table.cmd != "build" { // non-build cases end up inside the working dir -- need a global import path
		fpath = filepath.Join(curPkgPath, table.path)
	}
	args := []string{table.cmd, "-vm=" + pyvm, "-output=" + genPkgDir, "-package-prefix", table.pkgprefix}
	if table.extras != nil {
		args = append(args, table.extras...)
	}
	args = append(args, fpath)
	fmt.Printf("run cmd: %s\n", args)
	err = run(args)
	if err != nil {
		t.Fatalf("[%s:%s]: error running gopy-build: %v\n", pyvm, table.path, err)
	}

	// fmt.Printf("copying test.py\n")
	tstDir := genPkgDir
	if table.cmd != "build" {
		tstDir = filepath.Join(workdir, pkgNm)
	}
	if table.testdir != "" {
		tstDir = filepath.Join(workdir, table.testdir)
	}
	tstSrc := filepath.Join(filepath.Join(cwd, table.path), "test.py")
	tstDst := filepath.Join(tstDir, "test.py")
	err = copyCmd(tstSrc, tstDst)
	if err != nil {
		t.Fatalf("[%s:%s]: error copying 'test.py' from: %s to: %s: %v\n", pyvm, table.path, tstSrc, tstDst, err)
	}

	fmt.Printf("running %s test.py\n", pyvm)
	cmd := exec.Command(pyvm, "./test.py")
	cmd.Env = env
	cmd.Dir = tstDir
	cmd.Stdin = os.Stdin
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"[%s:%s]: error running python module: err=%v\n%v\n",
			pyvm, table.path,
			err,
			string(buf),
		)
	}

	var (
		got  = strings.Replace(string(buf), "\r\n", "\n", -1)
		want = strings.Replace(string(table.want), "\r\n", "\n", -1)
	)
	if !reflect.DeepEqual(got, want) {
		diffTxt := ""
		diffBin, diffErr := exec.LookPath("diff")
		if diffErr == nil {
			wantFile, wantErr := os.Create(filepath.Join(workdir, "want.txt"))
			if wantErr == nil {
				wantFile.Write([]byte(want))
				wantFile.Close()
			}
			gotFile, gotErr := os.Create(filepath.Join(workdir, "got.txt"))
			if gotErr == nil {
				gotFile.Write([]byte(got))
				gotFile.Close()
			}
			if gotErr == nil && wantErr == nil {
				cmd = exec.Command(diffBin, "-urN",
					wantFile.Name(),
					gotFile.Name(),
				)
				diff, _ := cmd.CombinedOutput()
				diffTxt = string(diff) + "\n"
			}
		}

		t.Fatalf("[%s:%s]: error running python module:\ngot:\n%s\n\nwant:\n%s\n[%s:%s] diff:\n%s",
			pyvm, table.path,
			got, want,
			pyvm, table.path,
			diffTxt,
		)
	}

}
