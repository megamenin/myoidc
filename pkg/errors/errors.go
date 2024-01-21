package errors

import (
	"fmt"
	"io"
)

type withMessage struct {
	cause   error
	message string
}

func (w *withMessage) Error() string {
	return w.message + ": " + w.cause.Error()
}

func (w *withMessage) Cause() error {
	return w.cause
}

func (w *withMessage) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.message)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}

type withCode struct {
	cause error
	code  int
}

func wrap(err error) error {
	if Is[StackTracer](err) {
		return err
	}
	return &based{
		cause: err,
		stack: callers(1),
	}

}

func (e withCode) Error() string {
	return e.cause.Error()
}

func (e withCode) Cause() error {
	return e.cause
}

func (e withCode) Code() int {
	return e.code
}

// based is an error that has a message and a stack, but no caller.
type based struct {
	cause error
	*stack
}

func (f *based) Error() string { return f.cause.Error() }

func (f *based) Cause() error {
	return f.cause
}

func (f *based) Unwrap() error {
	return f.cause
}

func (f *based) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, f.cause.Error())
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, f.cause.Error())
	case 'q':
		fmt.Fprintf(s, "%q", f.cause.Error())
	}
}

func (f based) StackTrace() string {
	return fmt.Sprintf("%+v", f.stack)
}

type withFields struct {
	cause  error
	fields map[string]interface{}
}

func (e withFields) Error() string {
	return e.cause.Error()
}

func (e withFields) Cause() error {
	return e.cause
}

func (e withFields) Fields() map[string]interface{} {
	return e.fields
}
