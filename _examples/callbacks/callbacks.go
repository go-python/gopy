// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package callbacks has Go functions that take Python callables, and call
// them before returning.
package callbacks

import (
	"fmt"
	"sync"
)

// Each calls fun for i in 0..n-1, with a label made from i.
func Each(n int, fun func(i int, label string)) {
	for i := 0; i < n; i++ {
		fun(i, fmt.Sprintf("item-%d", i))
	}
}

// Mixed calls fun with a bool, a float and an unsigned integer.
func Mixed(fun func(on bool, x float64, u uint8)) {
	fun(true, 1.5, 200)
	fun(false, -2.25, 7)
}

// Twice calls fun, which takes no arguments, two times.
func Twice(fun func()) {
	fun()
	fun()
}

// Counter counts how many times it has been visited.
type Counter struct {
	N int
}

// Visit calls fun with the counter itself, which arrives as a handle.
func (c *Counter) Visit(times int, fun func(c *Counter, n int)) {
	for i := 0; i < times; i++ {
		c.N++
		fun(c, c.N)
	}
}

// InGoroutine calls fun from another goroutine, and waits for it.
func InGoroutine(fun func(i int)) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fun(7)
	}()
	wg.Wait()
}
