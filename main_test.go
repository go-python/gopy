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
)

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

var testBackends = map[string]bool{}

func init() {
	if os.Getenv("GOPY_TRAVIS_CI") == "1" {
		log.Printf("Running in travis CI")
	}

	var disabled []string
	for _, be := range []struct {
		name   string
		vm     string
		module string
	}{
		{"py2", "python2", ""},
		{"py2-cffi", "python2", "cffi"},
		{"py3", "python3", ""},
		{"py3-cffi", "python3", "cffi"},
		{"pypy2-cffi", "pypy", "cffi"},
		{"pypy3-cffi", "pypy3", "cffi"},
	} {
		args := []string{"-c", ""}
		if be.module != "" {
			args[1] = "import " + be.module
		}
		log.Printf("checking testbackend: %q...", be.name)
		cmd := exec.Command(be.vm, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("disabling testbackend: %q, error: '%s'", be.name, err.Error())
			testBackends[be.name] = false
			disabled = append(disabled, be.name)
		} else {
			log.Printf("enabling testbackend: %q", be.name)
			testBackends[be.name] = true
		}
	}

	if len(disabled) > 0 {
		log.Printf("The following test backends are not available: %s",
			strings.Join(disabled, ", "))
		if os.Getenv("GOPY_TRAVIS_CI") == "1" {
			log.Fatalf("Not all backends available in travis CI")
		}
	}
}

type pkg struct {
	path string
	lang []string
	want []byte
}

func testPkgBackend(t *testing.T, lang, pycmd string, table pkg) {
	workdir, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		t.Fatalf("[%s:%s:%s]: could not create workdir: %v\n", lang, pycmd, table.path, err)
	}
	err = os.MkdirAll(workdir, 0644)
	if err != nil {
		t.Fatalf("[%s:%s:%s]: could not create workdir: %v\n", lang, pycmd, table.path, err)
	}
	defer os.RemoveAll(workdir)

	err = run([]string{"bind", "-lang=" + lang, "-output=" + workdir, "./" + table.path})
	if err != nil {
		t.Fatalf("[%s:%s:%s]: error running gopy-bind: %v\n", lang, pycmd, table.path, err)
	}

	cmd := exec.Command(
		"/bin/cp", "./"+table.path+"/test.py",
		filepath.Join(workdir, "test.py"),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("[%s:%s:%s]: error copying 'test.py': %v\n", lang, pycmd, table.path, err)
	}

	buf := new(bytes.Buffer)
	cmd = exec.Command(pycmd, "./test.py")
	cmd.Dir = workdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		t.Fatalf(
			"[%s:%s:%s]: error running python module: %v\n%v\n",
			lang, pycmd, table.path,
			err,
			string(buf.Bytes()),
		)
	}

	if !reflect.DeepEqual(string(buf.Bytes()), string(table.want)) {
		diffTxt := ""
		diffBin, diffErr := exec.LookPath("diff")
		if diffErr == nil {
			wantFile, wantErr := os.Create(filepath.Join(workdir, "want.txt"))
			if wantErr == nil {
				wantFile.Write(table.want)
				wantFile.Close()
			}
			gotFile, gotErr := os.Create(filepath.Join(workdir, "got.txt"))
			if gotErr == nil {
				gotFile.Write(buf.Bytes())
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

		t.Fatalf("[%s:%s:%s]: error running python module:\nwant:\n%s\n\ngot:\n%s\n[%s:%s:%s] diff:\n%s",
			lang, pycmd, table.path,
			string(table.want), string(buf.Bytes()),
			lang, pycmd, table.path,
			diffTxt,
		)
	}

}

var features = map[string][]string{
	// FIXME(sbinet): add "cffi" when go-python/gopy#130 and go-python/gopy#125
	// are fixed.
	"_examples/hi":        []string{"py2"},
	"_examples/funcs":     []string{"py2"},
	"_examples/simple":    []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/empty":     []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/named":     []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/structs":   []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/consts":    []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/vars":      []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/seqs":      []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/cgo":       []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/pyerrors":  []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/iface":     []string{},
	"_examples/pointers":  []string{},
	"_examples/arrays":    []string{"py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/slices":    []string{"py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/maps":      []string{"py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
	"_examples/gostrings": []string{"py2", "py2-cffi", "py3-cffi", "pypy2-cffi", "pypy3-cffi"},
}

func TestHi(t *testing.T) {
	t.Parallel()
	path := "_examples/hi"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`--- doc(hi)...
package hi exposes a few Go functions to be wrapped and used from Python.

--- hi.GetUniverse(): 42
--- hi.GetVersion(): 0.1
--- hi.GetDebug(): False
--- hi.SetDebug(true)
--- hi.GetDebug(): True
--- hi.SetDebug(false)
--- hi.GetDebug(): False
--- hi.GetAnon(): hi.Person{Name="<nobody>", Age=1}
--- new anon: hi.Person{Name="you", Age=24}
--- hi.SetAnon(hi.NewPerson('you', 24))...
--- hi.GetAnon(): hi.Person{Name="you", Age=24}
--- doc(hi.Hi)...
Hi() 

Hi prints hi from Go

--- hi.Hi()...
hi from go
--- doc(hi.Hello)...
Hello(str s) 

Hello prints a greeting from Go

--- hi.Hello('you')...
hello you from go
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
working...
worked for 2 hours
--- p.Work(24)...
working...
caught: can't work for 24 hours!
--- p.Salary(2): 20
--- p.Salary(24): caught: can't work for 24 hours!
--- Person.__init__
caught: invalid type for 'Name' attribute | err-type: <type 'exceptions.TypeError'>
caught: invalid type for 'Age' attribute | err-type: <type 'exceptions.TypeError'>
caught: Person.__init__ takes at most 2 argument(s) | err-type: <type 'exceptions.TypeError'>
hi.Person{Name="name", Age=0}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
--- hi.NewPerson('me', 666): hi.Person{Name="me", Age=666}
--- hi.NewPersonWithAge(666): hi.Person{Name="stranger", Age=666}
--- hi.NewActivePerson(4):
working...
worked for 4 hours
hi.Person{Name="", Age=0}
--- c = hi.Couple()...
hi.Couple{P1=hi.Person{Name="", Age=0}, P2=hi.Person{Name="", Age=0}}
--- c.P1: hi.Person{Name="", Age=0}
--- c: hi.Couple{P1=hi.Person{Name="tom", Age=5}, P2=hi.Person{Name="bob", Age=2}}
--- c = hi.NewCouple(tom, bob)...
hi.Couple{P1=hi.Person{Name="tom", Age=50}, P2=hi.Person{Name="bob", Age=41}}
hi.Couple{P1=hi.Person{Name="mom", Age=50}, P2=hi.Person{Name="bob", Age=51}}
--- Couple.__init__
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="", Age=0}}
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p2", Age=52}, P2=hi.Person{Name="p1", Age=42}}
caught: invalid type for 'P1' attribute | err-type: <type 'exceptions.TypeError'>
caught: invalid type for 'P1' attribute | err-type: <type 'exceptions.TypeError'>
caught: invalid type for 'P2' attribute | err-type: <type 'exceptions.TypeError'>
--- testing GC...
--- len(objs): 100000
--- len(vs): 100000
--- testing GC... [ok]
--- testing array...
arr: [2]int{1, 2}
len(arr): 2
arr[0]: 1
arr[1]: 2
arr[2]: caught: array index out of range
arr: [2]int{1, 42}
len(arr): 2
mem(arr): 2
--- testing slice...
slice: []int{1, 2}
len(slice): 2
slice[0]: 1
slice[1]: 2
slice[2]: caught: slice index out of range
slice: []int{1, 42}
len(slice): 2
mem(slice): 2
`),
	})

	// FIXME: Add to features when go-python/gopy#130 and
	// go-python/gopy#125 are fixed.
	testPkg(t, pkg{
		path: "_examples/hi",
		lang: []string{"py2-cffi"},
		want: []byte(`--- doc(hi)...
package hi exposes a few Go functions to be wrapped and used from Python.

--- hi.GetUniverse(): 42
--- hi.GetVersion(): 0.1
--- hi.GetDebug(): False
--- hi.SetDebug(true)
--- hi.GetDebug(): True
--- hi.SetDebug(false)
--- hi.GetDebug(): False
--- hi.GetAnon(): hi.Person{Name="<nobody>", Age=1}
--- new anon: hi.Person{Name="you", Age=24}
--- hi.SetAnon(hi.NewPerson('you', 24))...
--- hi.GetAnon(): hi.Person{Name="you", Age=24}
--- doc(hi.Hi)...
Hi() 

Hi prints hi from Go

--- hi.Hi()...
hi from go
--- doc(hi.Hello)...
Hello(str s) 

Hello prints a greeting from Go

--- hi.Hello('you')...
hello you from go
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
working...
worked for 2 hours
--- p.Work(24)...
working...
caught: can't work for 24 hours!
--- p.Salary(2): 20
--- p.Salary(24): caught: can't work for 24 hours!
--- Person.__init__
caught: invalid type for 'Name' attribute | err-type: <type 'exceptions.TypeError'>
caught: invalid type for 'Age' attribute | err-type: <type 'exceptions.TypeError'>
caught: Person.__init__ takes at most 2 argument(s) | err-type: <type 'exceptions.TypeError'>
hi.Person{Name="name", Age=0}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
hi.Person{Name="name", Age=42}
--- hi.NewPerson('me', 666): hi.Person{Name="me", Age=666}
--- hi.NewPersonWithAge(666): hi.Person{Name="stranger", Age=666}
--- hi.NewActivePerson(4):
working...
worked for 4 hours
hi.Person{Name="", Age=0}
--- c = hi.Couple()...
hi.Couple{P1=hi.Person{Name="", Age=0}, P2=hi.Person{Name="", Age=0}}
--- c.P1: hi.Person{Name="", Age=0}
--- c: hi.Couple{P1=hi.Person{Name="tom", Age=5}, P2=hi.Person{Name="bob", Age=2}}
--- c = hi.NewCouple(tom, bob)...
hi.Couple{P1=hi.Person{Name="tom", Age=50}, P2=hi.Person{Name="bob", Age=41}}
hi.Couple{P1=hi.Person{Name="mom", Age=50}, P2=hi.Person{Name="bob", Age=51}}
--- Couple.__init__
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="", Age=0}}
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p1", Age=42}, P2=hi.Person{Name="p2", Age=52}}
hi.Couple{P1=hi.Person{Name="p2", Age=52}, P2=hi.Person{Name="p1", Age=42}}
caught: 'int' object has no attribute 'cgopy' | err-type: <type 'exceptions.AttributeError'>
caught: 'int' object has no attribute 'cgopy' | err-type: <type 'exceptions.AttributeError'>
caught: 'int' object has no attribute 'cgopy' | err-type: <type 'exceptions.AttributeError'>
--- testing GC...
--- len(objs): 100000
--- len(vs): 100000
--- testing GC... [ok]
--- testing array...
arr: [2]int{1, 2}
len(arr): 2
arr[0]: 1
arr[1]: 2
arr[2]: caught: array index out of range
arr: [2]int{1, 42}
len(arr): 2
mem(arr): caught: cannot make memory view because object does not have the buffer interface
--- testing slice...
slice: []int{1, 2}
len(slice): 2
slice[0]: 1
slice[1]: 2
slice[2]: caught: slice index out of range
slice: []int{1, 42}
len(slice): 2
mem(slice): caught: cannot make memory view because object does not have the buffer interface
`),
	})
}

func TestBindFuncs(t *testing.T) {
	t.Parallel()
	path := "_examples/funcs"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`funcs.GetF1()...
calling F1
f1()= None
funcs.GetF2()...
calling F2
f2()= None
s1 = funcs.S1()...
s1.F1 = funcs.GetF2()...
calling F2
s1.F1() = None
s2 = funcs.S2()...
s2.F1 = funcs.GetF1()...
calling F1
s2.F1() = None
`),
	})
}

func TestBindSimple(t *testing.T) {
	t.Parallel()
	path := "_examples/simple"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`doc(pkg):
'simple is a simple package.\n'
pkg.Func()...
fct = pkg.Func...
fct()...
pkg.Add(1,2)= 3
pkg.Bool(True)= True
pkg.Bool(False)= False
pkg.Comp64Add((3+4j), (2+5j)) = (5+9j)
pkg.Comp128Add((3+4j), (2+5j)) = (5+9j)
`),
	})
}

func TestBindEmpty(t *testing.T) {
	t.Parallel()
	path := "_examples/empty"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`empty.init()... [CALLED]
doc(pkg):
'Package empty does not expose anything.\nWe may want to wrap and import it just for its side-effects.\n'
`),
	})
}

func TestBindPointers(t *testing.T) {
	t.Skip("not ready yet")
	t.Parallel()
	path := "_examples/pointers"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`s = pointers.S(2)
s = pointers.S{Value:2}
s.Value = 2
pointers.Inc(s)
==> go: s.Value==2
<== go: s.Value==3
s.Value = 3
`),
	})
}

func TestBindNamed(t *testing.T) {
	t.Parallel()
	path := "_examples/named"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`doc(named): 'package named tests various aspects of named types.\n'
doc(named.Float): ''
doc(named.Float.Value): 'Value() float\n\nValue returns a float32 value\n'
v = named.Float()
v = 0
v.Value() = 0.0
x = named.X()
x = 0
x.Value() = 0.0
x = named.XX()
x = 0
x.Value() = 0.0
x = named.XXX()
x = 0
x.Value() = 0.0
x = named.XXXX()
x = 0
x.Value() = 0.0
v = named.Float(42)
v = 42
v.Value() = 42.0
v = named.Float(42.0)
v = 42
v.Value() = 42.0
x = named.X(42)
x = 42
x.Value() = 42.0
x = named.XX(42)
x = 42
x.Value() = 42.0
x = named.XXX(42)
x = 42
x.Value() = 42.0
x = named.XXXX(42)
x = 42
x.Value() = 42.0
x = named.XXXX(42.0)
x = 42
x.Value() = 42.0
s = named.Str()
s = ""
s.Value() = ''
s = named.Str('string')
s = "string"
s.Value() = 'string'
arr = named.Array()
arr = named.Array{0, 0}
arr = named.Array([1,2])
arr = named.Array{1, 2}
arr = named.Array(range(10))
caught: Array.__init__ takes a sequence of size at most 2
arr = named.Array(xrange(2))
arr = named.Array{0, 1}
s = named.Slice()
s = named.Slice(nil)
s = named.Slice([1,2])
s = named.Slice{1, 2}
s = named.Slice(range(10))
s = named.Slice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s = named.Slice(xrange(10))
s = named.Slice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
`),
	})
}

func TestBindStructs(t *testing.T) {
	t.Parallel()
	path := "_examples/structs"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`s = structs.S()
s = structs.S{}
s.Init()
s.Upper('boo')= 'BOO'
s1 = structs.S1()
s1 = structs.S1{private:0}
caught error: 'S1' object has no attribute 'private'
s2 = structs.S2()
s2 = structs.S2{Public:0, private:0}
s2 = structs.S2(1)
s2 = structs.S2{Public:1, private:0}
caught error: S2.__init__ takes at most 1 argument(s)
s2 = structs.S2{Public:42, private:0}
s2.Public = 42
caught error: 'S2' object has no attribute 'private'
`),
	})
}

func TestBindConsts(t *testing.T) {
	t.Parallel()
	path := "_examples/consts"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`c1 = c1
c2 = 42
c3 = 666.666
c4 = c4
c5 = 42
c6 = 42
c7 = 666.666
k1 = 1
k2 = 2
`),
	})
}

func TestBindVars(t *testing.T) {
	t.Parallel()
	path := "_examples/vars"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`doc(vars):
''
doc(vars.GetV1()):
'returns vars.V1'
doc(vars.SetV1()):
'sets vars.V1'
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
vars.GetDoc() = 'A variable with some documentation'
doc of vars.GetDoc = 'returns vars.Doc\n\nDoc is a top-level string with some documentation attached.\n'
doc of vars.SetDoc = 'sets vars.Doc\n\nDoc is a top-level string with some documentation attached.\n'
`),
	})
}

func TestBindSeqs(t *testing.T) {
	t.Parallel()
	path := "_examples/seqs"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`doc(seqs): 'package seqs tests various aspects of sequence types.\n'
arr = seqs.Array(xrange(2))
arr = seqs.Array{0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
s = seqs.Slice()
s = seqs.Slice(nil)
s = seqs.Slice([1,2])
s = seqs.Slice{1, 2}
s = seqs.Slice(range(10))
s = seqs.Slice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s = seqs.Slice(xrange(10))
s = seqs.Slice{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s = seqs.Slice()
s = seqs.Slice(nil)
s += [1,2]
s = seqs.Slice{1, 2}
s += [10,20]
s = seqs.Slice{1, 2, 10, 20}
`),
	})
}

func TestBindInterfaces(t *testing.T) {
	t.Skip("not ready")
	t.Parallel()
	path := "_examples/iface"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`
`),
	})
}

func TestBindCgoPackage(t *testing.T) {
	t.Parallel()
	path := "_examples/cgo"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`cgo.doc: 'Package cgo tests bindings of CGo-based packages.\n'
cgo.Hi()= 'hi from go\n'
cgo.Hello(you)= 'hello you from go\n'
`),
	})
}

func TestPyErrors(t *testing.T) {
	t.Parallel()
	path := "_examples/pyerrors"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`Divide by zero.
pyerrors.Div(5, 2) = 2
`),
	})
}

func TestBuiltinArrays(t *testing.T) {
	t.Parallel()
	path := "_examples/arrays"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`Python list: [1, 2, 3, 4]
Go array:  [4]int{1, 2, 3, 4}
arrays.IntSum from Python list: 10
arrays.IntSum from Go array: 10
`),
	})
}

func TestBuiltinSlices(t *testing.T) {
	t.Parallel()
	path := "_examples/slices"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`Python list: [1, 2, 3, 4]
Go slice:  []int{1, 2, 3, 4}
slices.IntSum from Python list: 10
slices.IntSum from Go slice: 10
`),
	})
}

func TestBuiltinMaps(t *testing.T) {
	t.Parallel()
	path := "_examples/maps"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`maps.Sum from Go map: 8.0
maps.Sum from Python dictionary: 8.0
maps.Keys from Go map: []int{1, 2}
maps.Values from Go map: []float64{3, 5}
maps.Keys from Python dictionary: []int{1, 2}
maps.Values from Python dictionary: []float64{3, 5}
`),
	})
}

func TestBindStrings(t *testing.T) {
	t.Parallel()
	path := "_examples/gostrings"
	testPkg(t, pkg{
		path: path,
		lang: features[path],
		want: []byte(`S1 = S1
GetString() = MyString
`),
	})
}

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
