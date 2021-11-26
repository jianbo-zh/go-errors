package errors

import (
	"bytes"
)

type Context = map[string]interface{}

type Error struct {
	Inner error
	Msg   string
	Ctx   Context
}

func (err *Error) Error() string {
	if err.Inner == nil {
		return err.Msg
	}

	buff := bufferPool.Get().(*bytes.Buffer)
	buff.Reset()

	buff.WriteString(err.Msg)
	buff.Write(innerSeparator)
	buff.WriteString(err.Inner.Error())

	result := buff.String()
	bufferPool.Put(buff)

	return result
}

func (err *Error) Unwrap() error {
	return err.Inner
}

func (err *Error) Data() Context {
	return err.Ctx
}

func New(msg string, options ...Option) error {

	err := &Error{
		Msg: msg,
	}

	for _, option := range options {
		option(err)
	}

	return err
}

func Data(err error) Context {
	u, ok := err.(interface{ Data() Context })
	if !ok {
		return nil
	}
	return u.Data()
}
