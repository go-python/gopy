package multireturn

import (
	"fmt"
)

/////////////// No Return //////////////
func NoReturnFunc() {
}

/////////////// Single WithoutError Return //////////////
func SingleWithoutErrorFunc() int {
	return 100
}

/////////////// Single Str WithoutError Return //////////////
func SingleStrWithoutErrorFunc(vargs ...int) string {
	return "150"
}

/////////////// Single WithError Return //////////////
func SingleWithErrorFunc(throwError bool) error {
	if throwError {
		return fmt.Errorf("Error")
	} else {
		return nil
	}
}

/////////////// Double WithoutError Return //////////////
func DoubleWithoutErrorFunc1() (int, int) {
	return 200, 300
}
func DoubleWithoutErrorFunc2() (string, string) {
	return "200", "300"
}

/////////////// Double WithError Return //////////////
func DoubleWithErrorFunc(throwError bool) (string, error) {
	if throwError {
		return "400", fmt.Errorf("Error")
	} else {
		return "500", nil
	}
}

/////////////// Triple Returns Without Error //////////////
func TripleWithoutErrorFunc1(vargs ...int) (int, int, int) {
	return 600, 700, 800
}
func TripleWithoutErrorFunc2(vargs ...int) (int, string, int) {
	return 600, "700", 800
}

/////////////// Triple Returns With Error //////////////
func TripleWithErrorFunc(throwError bool) (int, int, error) {
	if throwError {
		return 900, 1000, fmt.Errorf("Error")
	} else {
		return 1100, 1200, nil
	}
}

/////////////// Triple Struct Returns With Error //////////////
type IntStrUct struct {
	P int
}

func NewIntStrUct(n int) IntStrUct {
	return IntStrUct{
		P: n,
	}
}

func TripleWithStructWithErrorFunc(throwError bool) (*IntStrUct, IntStrUct, error) {
	s1300 := IntStrUct{P: 1300}
	s1400 := IntStrUct{P: 1400}
	s1500 := IntStrUct{P: 1500}
	s1600 := IntStrUct{P: 1600}
	if throwError {
		return &s1300, s1400, fmt.Errorf("Error")
	} else {
		return &s1500, s1600, nil
	}
}

/////////////// Triple Interface Returns Without Error //////////////
type IntInterFace interface {
	Number() int
}

func (is *IntStrUct) Number() int {
	return is.P
}

func TripleWithInterfaceWithoutErrorFunc() (IntInterFace, IntStrUct, *IntStrUct) {
	i1700 := IntStrUct{P: 1700}
	s1800 := IntStrUct{P: 1800}
	s1900 := IntStrUct{P: 1900}

	return &i1700, s1800, &s1900
}

//// Function returning function /////
type FunctionType func(input int) int

func FunctionReturningFunction() FunctionType {
	return func(input int) int {
		return input
	}
}
