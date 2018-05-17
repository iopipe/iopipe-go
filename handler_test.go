package iopipe

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInvalidHandlers(t *testing.T) {

	testCases := []struct {
		name     string
		handler  interface{}
		expected error
	}{
		{
			name:     "nil handler",
			expected: errors.New("handler is nil"),
			handler:  nil,
		},
		{
			name:     "handler is not a function",
			expected: errors.New("handler kind struct is not func"),
			handler:  struct{}{},
		},
		{
			name:     "handler declares too many arguments",
			expected: errors.New("handlers may not take more than two arguments, but handler takes 3"),
			handler: func(n context.Context, x string, y string) error {
				return nil
			},
		},
		{
			name:     "two argument handler does not context as first argument",
			expected: errors.New("handler takes two arguments, but the first is not Context. got string"),
			handler: func(a string, x context.Context) error {
				return nil
			},
		},
		{
			name:     "handler returns too many values",
			expected: errors.New("handler may not return more than two values"),
			handler: func() (error, error, error) {
				return nil, nil, nil
			},
		},
		{
			name:     "handler returning two values does not declare error as the second return value",
			expected: errors.New("handler returns two values, but the second does not implement error"),
			handler: func() (error, string) {
				return nil, "hello"
			},
		},
		{
			name:     "handler returning a single value does not implement error",
			expected: errors.New("handler returns a single value, but it does not implement error"),
			handler: func() string {
				return "hello"
			},
		},
		{
			name:     "no return value should not result in error",
			expected: nil,
			handler: func() {
			},
		},
	}

	Convey("Invalid handlers result in errors", t, func() {
		for i, testCase := range testCases {
			Convey(fmt.Sprintf("testCase[%d] %s", i, testCase.name), func() {
				lambdaHandler := newHandler(testCase.handler)
				_, err := lambdaHandler(context.TODO(), make([]byte, 0))
				So(err, ShouldResemble, testCase.expected)
			})
		}
	})
}

type expected struct {
	val interface{}
	err error
}

func TestInvokes(t *testing.T) {
	hello := func(s string) string {
		return fmt.Sprintf("Hello %s!", s)
	}
	hellop := func(s *string) *string {
		v := hello(*s)
		return &v
	}

	testCases := []struct {
		name     string
		input    interface{}
		expected expected
		handler  interface{}
	}{
		{
			input:    `"Lambda"`,
			expected: expected{`Hello "Lambda"!`, nil},
			handler: func(name string) (string, error) {
				return hello(name), nil
			},
		},
		{
			input:    `"Lambda"`,
			expected: expected{`Hello "Lambda"!`, nil},
			handler: func(ctx context.Context, name string) (string, error) {
				return hello(name), nil
			},
		},
		{
			input: `"Lambda"`,
			expected: expected{
				func() *string {
					result := `Hello "Lambda"!`
					return &result
				}(),
				nil},
			handler: func(name *string) (*string, error) {
				return hellop(name), nil
			},
		},
		{
			input: `"Lambda"`,
			expected: expected{
				func() *string {
					result := `Hello "Lambda"!`
					return &result
				}(),
				nil},
			handler: func(ctx context.Context, name *string) (*string, error) {
				return hellop(name), nil
			},
		},
		{
			input:    `"Lambda"`,
			expected: expected{"", errors.New("bad stuff")},
			handler: func() error {
				return errors.New("bad stuff")
			},
		},
		{
			input:    `"Lambda"`,
			expected: expected{"", errors.New("bad stuff")},
			handler: func() (interface{}, error) {
				return nil, errors.New("bad stuff")
			},
		},
		{
			input:    `"Lambda"`,
			expected: expected{"", errors.New("bad stuff")},
			handler: func(e interface{}) (interface{}, error) {
				return nil, errors.New("bad stuff")
			},
		},
		{
			input:    `"Lambda"`,
			expected: expected{"", errors.New("bad stuff")},
			handler: func(ctx context.Context, e interface{}) (interface{}, error) {
				return nil, errors.New("bad stuff")
			},
		},
		{
			name:     "basic input struct serialization",
			input:    struct{ Custom int }{9001},
			expected: expected{9001, nil},
			handler: func(event struct{ Custom int }) (int, error) {
				return event.Custom, nil
			},
		},
		{
			name:     "basic output struct serialization",
			input:    9001,
			expected: expected{struct{ Number int }{9001}, nil},
			handler: func(event int) (struct{ Number int }, error) {
				return struct{ Number int }{event}, nil
			},
		},
	}

	Convey("Valid handlers work", t, func() {
		for i, testCase := range testCases {
			Convey(fmt.Sprintf("testCase[%d] %s", i, testCase.name), func() {
				lambdaHandler := newHandler(testCase.handler)
				response, err := lambdaHandler(context.TODO(), testCase.input)
				if testCase.expected.err != nil {
					So(err, ShouldResemble, testCase.expected.err)
				} else {
					So(err, ShouldBeNil)
					So(response, ShouldResemble, testCase.expected.val)
				}
			})
		}
	})
}

func TestInvalidJsonInput(t *testing.T) {
	Convey("json: unsupported type", t, func() {
		lambdaHandler := newHandler(func(s string) error { return nil })
		_, err := lambdaHandler(context.TODO(), make(chan int))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "json: unsupported type: chan int")
	})

}
