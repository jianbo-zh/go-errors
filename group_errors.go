package errors

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.uber.org/atomic"
)

type errorGroup interface {
	// Returns a slice containing the underlying list of errors.
	//
	// This slice MUST NOT be modified by the caller.
	Errors() []error
}

// MultiError is an error that holds one or more errors.
//
// An instance of this is guaranteed to be non-empty and flattened. That is,
// none of the errors inside MultiError are other MultiErrors.
//
// MultiError formats to a semi-colon delimited list of error messages with
// %v and with a more readable multi-line format with %+v.
type MultiError struct {
	copyNeeded atomic.Bool
	errors     []error
}

var _ errorGroup = (*MultiError)(nil)

// Errors returns the list of underlying errors.
//
// This slice MUST NOT be modified.
func (merr *MultiError) Errors() []error {
	if merr == nil {
		return nil
	}
	return merr.errors
}

func (merr *MultiError) Error() string {

	if len(merr.errors) == 0 {
		return ""
	}

	buff := bufferPool.Get().(*bytes.Buffer)
	buff.Reset()

	merr.writeSingleline(buff)

	result := buff.String()
	bufferPool.Put(buff)

	return result
}

func (merr *MultiError) As(target interface{}) bool {
	for _, err := range merr.Errors() {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func (merr *MultiError) Is(target error) bool {
	for _, err := range merr.Errors() {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (merr *MultiError) Format(f fmt.State, c rune) {
	if c == 'v' && f.Flag('+') {
		merr.writeMultiline(f)
	} else {
		merr.writeSingleline(f)
	}
}

func (merr *MultiError) writeSingleline(w io.Writer) {
	first := true
	for _, item := range merr.errors {
		if first {
			first = false
		} else {
			w.Write(groupSeparator)
		}
		io.WriteString(w, item.Error())
	}
}

func (merr *MultiError) writeMultiline(w io.Writer) {
	w.Write(multilinePrefix)
	for _, item := range merr.errors {
		w.Write(multilineSeparator)
		writePrefixLine(w, multilineIndent, fmt.Sprintf("%+v", item))
	}
}

func Combine(errors ...error) error {
	return fromSlice(errors)
}

// Append appends the given errors together. Either value may be nil.
//
// This function is a specialization of Combine for the common case where
// there are only two errors.
//
// 	err = multierr.Append(reader.Close(), writer.Close())
//
// The following pattern may also be used to record failure of deferred
// operations without losing information about the original error.
//
// 	func doSomething(..) (err error) {
// 		f := acquireResource()
// 		defer func() {
// 			err = multierr.Append(err, f.Close())
// 		}()
func Append(left error, right error) error {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	}

	if _, ok := right.(*MultiError); !ok {
		if l, ok := left.(*MultiError); ok && !l.copyNeeded.Swap(true) {
			// Common case where the error on the left is constantly being
			// appended to.
			errs := append(l.errors, right)
			return &MultiError{errors: errs}
		} else if !ok {
			// Both errors are single errors.
			return &MultiError{errors: []error{left, right}}
		}
	}

	// Either right or both, left and right, are MultiErrors. Rely on usual
	// expensive logic.
	errors := [2]error{left, right}
	return fromSlice(errors[0:])
}

func Errors(err error) []error {
	if err == nil {
		return nil
	}

	// Note that we're casting to MultiError, not errorGroup. Our contract is
	// that returned errors MAY implement errorGroup. Errors, however, only
	// has special behavior for multierr-specific error objects.
	//
	// This behavior can be expanded in the future but I think it's prudent to
	// start with as little as possible in terms of contract and possibility
	// of misuse.
	eg, ok := err.(errorGroup)
	if !ok {
		return []error{err}
	}

	errors := eg.Errors()
	result := make([]error, len(errors))

	copy(result, errors)

	return result
}

// fromSlice converts the given list of errors into a single error.
func fromSlice(errors []error) error {
	res := inspect(errors)
	switch res.Count {
	case 0:
		return nil
	case 1:
		// only one non-nil entry
		return errors[res.FirstErrorIdx]
	case len(errors):
		if !res.ContainsMultiError {
			// already flat
			return &MultiError{errors: errors}
		}
	}

	nonNilErrs := make([]error, 0, res.Capacity)
	for _, err := range errors[res.FirstErrorIdx:] {
		if err == nil {
			continue
		}

		if nested, ok := err.(*MultiError); ok {
			nonNilErrs = append(nonNilErrs, nested.errors...)
		} else {
			nonNilErrs = append(nonNilErrs, err)
		}
	}

	return &MultiError{errors: nonNilErrs}
}

// Writes s to the writer with the given prefix added before each line after
// the first.
func writePrefixLine(w io.Writer, prefix []byte, s string) {
	first := true
	for len(s) > 0 {
		if first {
			first = false
		} else {
			w.Write(prefix)
		}

		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			idx = len(s) - 1
		}

		io.WriteString(w, s[:idx+1])
		s = s[idx+1:]
	}
}

type inspectResult struct {
	// Number of top-level non-nil errors
	Count int

	// Total number of errors including MultiErrors
	Capacity int

	// Index of the first non-nil error in the list. Value is meaningless if
	// Count is zero.
	FirstErrorIdx int

	// Whether the list contains at least one MultiError
	ContainsMultiError bool
}

// Inspects the given slice of errors so that we can efficiently allocate
// space for it.
func inspect(errors []error) (res inspectResult) {
	first := true
	for i, err := range errors {
		if err == nil {
			continue
		}

		res.Count++
		if first {
			first = false
			res.FirstErrorIdx = i
		}

		if merr, ok := err.(*MultiError); ok {
			res.Capacity += len(merr.errors)
			res.ContainsMultiError = true
		} else {
			res.Capacity++
		}
	}
	return
}
