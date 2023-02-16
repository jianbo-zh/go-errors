package errors

import (
	"bytes"
	"errors"
	"fmt"
)

type MapData = map[string]interface{}

type Error struct {
	Inner    error
	Code     int
	Messge   string
	Layer    string
	Category string
	Payload  MapData
}

func (err *Error) Error() string {
	if err.Inner == nil {
		return err.Messge
	}

	buff := bufferPool.Get().(*bytes.Buffer)
	buff.Reset()

	buff.WriteString(err.Messge)
	buff.Write(innerSeparator)
	buff.WriteString(err.Inner.Error())

	result := buff.String()
	bufferPool.Put(buff)

	return result
}

func (err *Error) Unwrap() error {
	return err.Inner
}

func (err *Error) With(options ...Option) *Error {
	for _, option := range options {
		option(err)
	}
	return err
}

func New(msg string) *Error {
	err := &Error{
		Messge: msg,
	}
	return err
}

func Newf(format string, a ...interface{}) *Error {
	err := &Error{
		Messge: fmt.Sprintf(format, a...),
	}
	return err
}

func Newc(code int, msg string) *Error {
	err := &Error{
		Code:   code,
		Messge: msg,
	}
	return err
}

func Newcf(code int, format string, a ...interface{}) *Error {
	err := &Error{
		Code:   code,
		Messge: fmt.Sprintf(format, a...),
	}
	return err
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
