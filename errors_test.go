package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err1 := New("world")
	err := New("hello", WithError(err1))

	fmt.Printf("%+v", err)
	fmt.Printf("err is err1 %t", errors.Is(err, err1))
}

func TestCombine(t *testing.T) {
	err1 := New("hello", WithError(errors.New("key world")))
	err2 := New("world")
	err3 := New("hahaha")

	err := Combine(err1, err2, err3)

	fmt.Printf("%+v", err)

	fmt.Printf("err is err2 %t", errors.Is(err, err2))
}

func TestAppend(t *testing.T) {
	err1 := New("hello", WithError(errors.New("key world")))
	err2 := New("world")

	fmt.Printf("%+v", Append(err1, err2))
}
