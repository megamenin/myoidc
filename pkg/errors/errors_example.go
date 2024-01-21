package errors

import (
	"fmt"
)

func ExampleAs() {
	err := test1()
	fmt.Printf("message: %s\n", err)
	var tracer StackTracer
	if As(err, &tracer) {
		fmt.Printf("trace: %s\n", tracer.StackTrace())
	}
	var coder Coder
	if As(err, &coder) {
		fmt.Printf("code: %d\n", coder.Code())
	}
	var fielder Fielder
	if As(err, &fielder) {
		fmt.Printf("%+v\n", fielder.Fields())
	}
}

func test1() error {
	err := test2()
	err = WithField(err, "filed2", 22)
	return err
}

func test2() error {
	return WithCode(test3(), 2)
}

func test3() error {
	return test4()
}

func test4() error {
	return WithCode(test5(), 5)
}

func test5() error {
	err := test6()
	err = Wrap(err, "test 5 error")
	err = WithField(err, "filed5", 55555)
	return err
}

func test6() error {
	return fmt.Errorf("my message")
}
