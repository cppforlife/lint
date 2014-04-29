package errorassignment

import (
	"errors"
	"fmt"
)

func testNotAssigned() {
	// External library
	fmt.Printf("hello")

	// Single return value not assigned
	testSe()

	// Multiple return values (1 error type) not assigned
	testMe()

	// Multiple return values (2 error types) not assigned
	testMe2()
}

func testNotUsed() {
	// External library
	_, _ = fmt.Printf("hello")

	// Single return value not used
	_ = testSe()

	// Multiple return values (1 error type) not used
	_, _ = testMe()

	// Multiple return values (2 error types) not used
	_, err, _ := testMe2()

	println(err.Error())
}

func testSe() error {
	return errors.New("desc")
}

func testMe() (int, error) {
	return 1, errors.New("desc")
}

func testMe2() (int, error, error) {
	return 1, errors.New("desc"), errors.New("desc")
}
