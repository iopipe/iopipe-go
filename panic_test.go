package iopipe

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"fmt"
	"runtime"
)

func testFunctionThatPanics() {
	panic(fmt.Errorf("i am a throwing function"))
}

func TestNewHandlerError(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Given a parent function that recovers from panics, stores the error using NewHandlerError, and wraps a panicking child function", t, func() {
		var panicErr *handlerError

		parentWithPanickyChild := func() {
			defer func() {
				if err := recover(); err != nil {
					panicErr = NewHandlerError(err, true)
				}
			}()

			testFunctionThatPanics()
		}

		Convey("The child function should panic", func() {
			So(testFunctionThatPanics, ShouldPanic)
		})

		Convey("The parent function should not panic", func() {
			So(parentWithPanickyChild, ShouldNotPanic)
		})

		Convey("When the parent function is called", func() {
			parentWithPanickyChild()

			Convey("The first stack frame in NewHandlerError refers to the panic inside the child function", func() {
				So(*panicErr.StackTrace[0], ShouldResemble, panicErrorStackFrame{
					Path:     "github.com/iopipe/iopipe-go/panic_test.go",
					Line:     11,
					Function: "testFunctionThatPanics",
				})
			})
		})
	})
}

func assertPanicMessage(t *testing.T, panicFunc func(), expectedMessage string) {
	defer func() {
		if err := recover(); err != nil {
			panicInfo := getPanicInfo(err, 0)
			So(panicInfo, ShouldNotBeNil)
			So(panicInfo.Message, ShouldNotBeNil)
			So(panicInfo.Message, ShouldEqual, expectedMessage)
			So(panicInfo.StackTrace, ShouldNotBeNil)
		}
	}()

	panicFunc()
	t.Errorf("Should have exited due to panic")
}

func TestPanicFormattingStringValue(t *testing.T) {
	Convey("Panic formatting is accurate for String message values", t, func() {
		assertPanicMessage(t, func() { panic("Panic time!") }, "Panic time!")
	})
}

func TestPanicFormattingIntValue(t *testing.T) {
	Convey("Panic formatting is accurate for Int message values", t, func() {
		assertPanicMessage(t, func() { panic(1234) }, "1234")
	})
}

type CustomError struct{}
func (e CustomError) Error() string { return fmt.Sprintf("Something bad happened!") }
func TestPanicFormattingCustomError(t *testing.T) {
	Convey("Panic formatting is accurate for Customer Error type messages", t, func() {
		customError := &CustomError{}
		assertPanicMessage(t, func() { panic(customError) }, customError.Error())
	})
}

func TestFormatFrame(t *testing.T) {
	Convey("FormatFrame appropriately formats a variety of runtime.Frame examples", t, func() {
		var tests = []struct {
			inputPath        string
			inputLine        int32
			inputFunction    string
			expectedPath     string
			expectedLine     int32
			expectedFunction string
		}{
			{
				inputPath:        "/Volumes/Unix/workspace/LambdaGoLang/src/GoAmzn-Github-Aws-AwsLambdaGo/src/github.com/aws/aws-lambda-go/lambda/panic_test.go",
				inputLine:        42,
				inputFunction:    "github.com/aws/aws-lambda-go/lambda.printStack",
				expectedPath:     "github.com/aws/aws-lambda-go/lambda/panic_test.go",
				expectedLine:     42,
				expectedFunction: "printStack",
			},
			{
				inputPath:        "/home/user/src/pkg/sub/file.go",
				inputLine:        42,
				inputFunction:    "pkg/sub.Type.Method",
				expectedPath:     "pkg/sub/file.go",
				expectedLine:     42,
				expectedFunction: "Type.Method",
			},
			{
				inputPath:        "/home/user/src/pkg/sub/sub2/file.go",
				inputLine:        42,
				inputFunction:    "pkg/sub/sub2.Type.Method",
				expectedPath:     "pkg/sub/sub2/file.go",
				expectedLine:     42,
				expectedFunction: "Type.Method",
			},
			{
				inputPath:        "/home/user/src/pkg/file.go",
				inputLine:        101,
				inputFunction:    "pkg.Type.Method",
				expectedPath:     "pkg/file.go",
				expectedLine:     101,
				expectedFunction: "Type.Method",
			},
		}

		for _, test := range tests {
			inputFrame := runtime.Frame{
				File:     test.inputPath,
				Line:     int(test.inputLine),
				Function: test.inputFunction,
			}

			actual := formatFrame(inputFrame)

			Convey(fmt.Sprintf(`Frame with inputPath "%s" should be formatted correctly`, test.inputPath), func() {
				So(actual.Path, ShouldEqual, test.expectedPath)
				So(actual.Line, ShouldEqual, test.expectedLine)
				So(actual.Function, ShouldEqual, test.expectedFunction)
			})
		}
	})
}

func TestRuntimeStackTrace(t *testing.T) {
	// implementing the test in the inner function to simulate an
	// additional stack frame that would exist in real life due to the
	// defer function.
	testRuntimeStackTrace(t)
}

func testRuntimeStackTrace(t *testing.T) {
	Convey("Runtime stack trace's first element is where the panic happens", t, func() {
		const framesToHide = framesToPanicInfo
		panicInfo := getPanicInfo("Panic time!", framesToHide)

		So(panicInfo , ShouldNotBeNil)
		So(panicInfo.StackTrace , ShouldNotBeNil)
		So(len(panicInfo.StackTrace), ShouldBeGreaterThan, 0)

		frame := panicInfo.StackTrace[0]

		So(frame.Path, ShouldEqual, "github.com/iopipe/iopipe-go/panic_test.go")
		So(frame.Line, ShouldEqual, 160)
		So(frame.Function, ShouldEqual, "testRuntimeStackTrace.func1")
	})
}
