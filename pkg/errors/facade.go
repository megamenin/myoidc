package errors

import "fmt"

func New(message string) error {
	return &based{
		cause: fmt.Errorf(message),
		stack: callers(),
	}
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:   wrap(err),
		message: message,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return err
	}
	return &withMessage{
		cause:   wrap(err),
		message: fmt.Sprintf(format, args...),
	}
}

func Error(message string) error {
	return &based{
		cause: fmt.Errorf(message),
		stack: callers(),
	}
}

func Errorf(format string, args ...interface{}) error {
	return &based{
		cause: fmt.Errorf(format, args...),
		stack: callers(),
	}
}

func WithCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withCode{
		cause: wrap(err),
		code:  code,
	}
}

func WithField(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}
	var fielder Fielder
	if As(err, &fielder) {
		fields := fielder.Fields()
		fields[key] = value
		return err
	}
	return &withFields{
		cause: wrap(err),
		fields: map[string]interface{}{
			key: value,
		},
	}
}

func Is[T error](err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(T); ok {
		return true
	}
	if e, ok := err.(Causer); ok {
		return Is[T](e.Cause())
	}
	return false
}

func As[T error](err error, dist *T) bool {
	if dist == nil {
		return false
	}
	if Is[T](err) {
		*dist = Unwrap[T](err)
		return true
	}
	return false
}

func Unwrap[T error](err error) T {
	var res T
	if err == nil {
		return res
	}
	if e, ok := err.(T); ok {
		return e
	}
	if e, ok := err.(Causer); ok {
		return Unwrap[T](e.Cause())
	}
	return res
}

func IsStackTracer(err error) bool {
	return Is[StackTracer](err)
}

func IsCoder(err error) bool {
	return Is[Coder](err)
}

func GetStackTrace(err error) string {
	e := Unwrap[StackTracer](err)
	if e != nil {
		return e.StackTrace()
	}
	return ""
}

func GetErrCode(err error) int {
	e := Unwrap[Coder](err)
	if e != nil {
		return e.Code()
	}
	return 0
}

func GetFields(err error) map[string]interface{} {
	e := Unwrap[Fielder](err)
	if e != nil {
		return e.Fields()
	}
	return map[string]interface{}{}
}
