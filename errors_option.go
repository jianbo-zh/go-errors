package errors

type Option func(err *Error)

func Playload(playload MapData) Option {
	return func(err *Error) {
		err.Playload = playload
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
