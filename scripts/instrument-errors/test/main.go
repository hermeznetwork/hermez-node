//nolint
package main

import (
	"fmt"
	"io"

	"github.com/ztrue/tracerr"
)

func badFn1() error {
	err := fmt.Errorf("error")
	return err
}

func badFn2() (int, error) {
	err := fmt.Errorf("error")
	v := 1
	return v, err
}

func badFn3() error {
	return fmt.Errorf("error")
}

func badFn4() (*int, error) {
	return nil, fmt.Errorf("error")
}

func goodFn1() error {
	err := fmt.Errorf("error")
	return tracerr.Wrap(err)
}

func goodFn2() error {
	return tracerr.Wrap(fmt.Errorf("error"))
}

func badFnCheck1() {
	err := fmt.Errorf("error")
	if err == io.EOF {
		fmt.Println("Foo")
	}
	if err != io.EOF {
		fmt.Println("Bar")
	}
}

func goofFnCheck1() {
	err := fmt.Errorf("error")
	if tracerr.Unwrap(err) == io.EOF {
		fmt.Println("Foo")
	}
	if tracerr.Unwrap(err) != io.EOF {
		fmt.Println("Bar")
	}
}
