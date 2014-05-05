## Usage

```
cd $GOPATH
go get github.com/cppforlife/lint
go install github.com/cppforlife/lint
./bin/lint github.com/cppforlife/lint
```

Example output of linting itself (test cases errors):

```
Looking at package "github.com/cppforlife/lint/testcase/packagedirname_test"
Looking at package "github.com/cppforlife/lint/testcase/packagedirname"

-- /tmp/go/src/github.com/cppforlife/lint/testcase/packagedirname/main_test.go
main_test.go:1:1 Test package name should match directory name with _text suffix
	dirName = packagedirname
	package : pkg_test -> packagedirname_test

-- /tmp/go/src/github.com/cppforlife/lint/testcase/packagedirname/main.go
main.go:1:1 Package name should match directory name
	dirName = packagedirname
	package : pkg -> packagedirname

Looking at package "github.com/cppforlife/lint/testcase/testpackagesuffix"

-- /tmp/go/src/github.com/cppforlife/lint/testcase/testpackagesuffix/main_test.go
main_test.go:2:1 Test file should be in a corresponding test package
	fileName = main_test.go
	package : testpackagesuffix -> testpackagesuffix_test

Looking at package "github.com/cppforlife/lint/testcase/packagedirnamemain_test"
Looking at package "github.com/cppforlife/lint/testcase/packagedirnamemain"
Looking at package "github.com/cppforlife/lint/testcase/errorassignment"

-- /tmp/go/src/github.com/cppforlife/lint/testcase/errorassignment/main.go
main.go:10:6 Return value of type error should be assigned and used
	func = func fmt.Printf(format string, a ...interface{}) (n int, err error)
main.go:13:2 Return value of type error should be assigned and used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testSe() error
main.go:16:2 Return value of type error should be assigned and used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testMe() (int, error)
main.go:19:2 Return value of type error should be assigned and used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testMe2() (int, error, error)
main.go:19:2 Return value of type error should be assigned and used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testMe2() (int, error, error)
main.go:24:5 Return value of type error should be used
	func = func fmt.Printf(format string, a ...interface{}) (n int, err error)
main.go:27:2 Return value of type error should be used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testSe() error
main.go:30:5 Return value of type error should be used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testMe() (int, error)
main.go:33:10 Return value of type error should be used
	func = func github.com/cppforlife/lint/testcase/errorassignment.testMe2() (int, error, error)

Looking at package "github.com/cppforlife/lint/check"
Looking at package "github.com/cppforlife/lint/linter"
Looking at package "github.com/cppforlife/lint/testcase"
Looking at package "github.com/cppforlife/lint_test"
Looking at package "github.com/cppforlife/lint"
````

## Todo

- Add `https://github.com/golang/lint` as a check
- Check for unused/unassigned errors in defer and go stmts

## Notes

- http://godoc.org/go/ast
- http://godoc.org/code.google.com/p/go.tools/go/loader
- http://godoc.org/code.google.com/p/go.tools/go/types
