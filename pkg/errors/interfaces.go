package errors

type Causer interface {
	error
	Cause() error
}

type StackTracer interface {
	Causer
	StackTrace() string
}

type Coder interface {
	Causer
	Code() int
}

type Fielder interface {
	Causer
	Fields() map[string]interface{}
}
