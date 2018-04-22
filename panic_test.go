package iopipe

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"fmt"
)

func TestNewPanicError(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Given a function that wraps a function that panics", t, func() {
		var panicErr *handlerError

		recoverAndPanicWrapper := func(f func()) {
			defer func() {
				if err := recover(); err != nil {
					panic(err)
				}
			}()

			f()
			Convey("Impossible!", func() {
				So(1, ShouldEqual, 2)
			})
		}

		f := func() {
			panic(fmt.Errorf("i am a throwing function"))
		}

		main := func() {
			defer func() {
				if err := recover(); err != nil {
					panicErr = NewHandlerError(err, true)
				}
			}()

			recoverAndPanicWrapper(f)
			fmt.Println("Returned normally from f.")
		}

		Convey("When the wrapper function is called inside another function", func() {
			main()

			Convey("The line of first element in the stack trace is where the original panic occurs", func() {
				So(panicErr.StackTrace[0].Line, ShouldEqual, 29)
			})

			Convey("The path of first element in the stack trace is the path to this test file", func() {
				So(panicErr.StackTrace[0].Path, ShouldEqual, "/go/src/github.com/iopipe/iopipe-go/panic_test.go")
			})
		})
	})
}
