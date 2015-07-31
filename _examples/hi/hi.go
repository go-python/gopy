// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package hi exposes a few Go functions to be wrapped and used from Python.
package hi

import (
	"fmt"
)

const (
	Version  = "0.1" // Version of this package
	Universe = 42    // Universe is the fundamental constant of everything
)

var (
	Debug = false                            // Debug switches between debug and prod
	Anon  = Person{Age: 1, Name: "<nobody>"} // Anon is a default anonymous person
)

// Hi prints hi from Go
func Hi() {
	fmt.Printf("hi from go\n")
}

// Hello prints a greeting from Go
func Hello(s string) {
	fmt.Printf("hello %s from go\n", s)
}

// Concat concatenates two strings together and returns the resulting string.
func Concat(s1, s2 string) string {
	return s1 + s2
}

// Add returns the sum of its arguments.
func Add(i, j int) int {
	return i + j
}

// Person is a simple struct
type Person struct {
	Name string
	Age  int
}

// NewPerson creates a new Person value
func NewPerson(name string, age int) Person {
	return Person{
		Name: name,
		Age:  age,
	}
}

// NewPersonWithAge creates a new Person with a specific age
func NewPersonWithAge(age int) Person {
	return Person{
		Name: "stranger",
		Age:  age,
	}
}

// NewActivePerson creates a new Person with a certain amount of work done.
func NewActivePerson(h int) (Person, error) {
	var p Person
	err := p.Work(h)
	return p, err
}

func (p Person) String() string {
	return fmt.Sprintf("hi.Person{Name=%q, Age=%d}", p.Name, p.Age)
}

// Greet sends greetings
func (p *Person) Greet() string {
	return p.greet()
}

// greet sends greetings
func (p *Person) greet() string {
	return fmt.Sprintf("Hello, I am %s", p.Name)
}

// Work makes a Person go to work for h hours
func (p *Person) Work(h int) error {
	fmt.Printf("working...\n")
	if h > 7 {
		return fmt.Errorf("can't work for %d hours!", h)
	}
	fmt.Printf("worked for %d hours\n", h)
	return nil
}

// Salary returns the expected gains after h hours of work
func (p *Person) Salary(h int) (int, error) {
	if h > 7 {
		return 0, fmt.Errorf("can't work for %d hours!", h)
	}
	return h * 10, nil
}

// Couple is a pair of persons
type Couple struct {
	P1 Person
	P2 Person
}

// FIXME(sbinet) -- expose!
// NewCouple returns a new couple made of the p1 and p2 persons.
func newCouple(p1, p2 Person) Couple {
	return Couple{
		P1: p1,
		P2: p2,
	}
}

func (c *Couple) String() string {
	return fmt.Sprintf("hi.Couple{P1=%v, P2=%v}", c.P1, c.P2)
}
