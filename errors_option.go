package errors

type Option func(err *Error)

func WithContext(data Context) Option {
	return func(err *Error) {
		err.Ctx = data
	}
}

func WithError(inner error) Option {
	return func(err *Error) {
		err.Inner = inner
	}
}
