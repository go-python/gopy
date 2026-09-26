// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package callbacks has Go functions that take Python callables, and call
// them before returning.
package callbacks

import (
	"fmt"
	"sync"
	"time"
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

// Describe calls fun with a string and a fmt.Stringer, as interface{} values.
// They arrive as strings, made by fmt.Sprintf("%s", v).
func Describe(fun func(v interface{})) {
	fun("a string")
	fun(1500 * time.Millisecond)
}

// Count returns how many of 0..n-1 keep says yes to.
func Count(n int, keep func(i int) bool) int {
	total := 0
	for i := 0; i < n; i++ {
		if keep(i) {
			total++
		}
	}
	return total
}

// Sum adds up what val returns for 0..n-1.
func Sum(n int, val func(i int) int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += val(i)
	}
	return total
}

// Widest returns the largest of what size returns for 0..n-1.
func Widest(n int, size func(i int) uint) uint {
	var widest uint
	for i := 0; i < n; i++ {
		if w := size(i); w > widest {
			widest = w
		}
	}
	return widest
}

// Apply returns f(x).
func Apply(x float64, f func(x float64) float64) float64 {
	return f(x)
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

// Check calls fun with the counter itself, and reports what it answered.
func (c *Counter) Check(fun func(c *Counter, n int) bool) bool {
	return fun(c, c.N)
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
