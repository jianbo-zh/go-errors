package errors

type Option func(err *Error)

func Payload(payload MapData) Option {
	return func(err *Error) {
		err.Payload = payload
	}
}

func Inner(inner error) Option {
	return func(err *Error) {
		err.Inner = inner
	}
}

func Layer(layer string) Option {
	return func(err *Error) {
		err.Layer = layer
	}
}

func Category(category string) Option {
	return func(err *Error) {
		err.Category = category
	}
}
