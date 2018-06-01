package iopipe

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAgent_NewAgent(t *testing.T) {
	Convey("An agent created with empty Config should have default configuration", t, func() {
		agentWithDefaultConfig := NewAgent(Config{})

		Convey("Agent should be enabled", func() {
			So(agentWithDefaultConfig.Plugins, ShouldBeEmpty)
		})

		Convey("Plugins should be empty", func() {
			So(agentWithDefaultConfig.Plugins, ShouldBeEmpty)
		})

		Convey("TimeoutWindow should be 150 ms", func() {
			So(*agentWithDefaultConfig.TimeoutWindow, ShouldEqual, 150*time.Millisecond)
		})
	})

	Convey("An agent created with custom Config should assume that configuration", t, func() {

		var tests = []struct {
			inputToken         string
			inputEnabled       bool
			inputTimeoutWindow time.Duration
		}{
			{
				inputToken:         "",
				inputEnabled:       false,
				inputTimeoutWindow: 0 * time.Millisecond,
			},
			{
				inputToken:         "non-empty-value",
				inputEnabled:       true,
				inputTimeoutWindow: 20 * time.Millisecond,
			},
		}

		for _, test := range tests {
			inputConfig := Config{
				Token:         &test.inputToken,
				Enabled:       &test.inputEnabled,
				TimeoutWindow: &test.inputTimeoutWindow,
			}

			agentWithCustomConfig := NewAgent(inputConfig)

			Convey(fmt.Sprintf(`Config with inputToken "%s" should be mapped corectly in the agent instance`, test.inputToken), func() {
				So(*agentWithCustomConfig.Token, ShouldEqual, test.inputToken)
				So(*agentWithCustomConfig.Enabled, ShouldEqual, test.inputEnabled)
				So(*agentWithCustomConfig.TimeoutWindow, ShouldEqual, test.inputTimeoutWindow)
				So(agentWithCustomConfig.Plugins, ShouldBeEmpty)
			})
		}

	})
}

func TestAgent_WrapHandler(t *testing.T) {
	Convey("Returns original handler when Config.Enabled = false", t, func() {
		handler := func() {}
		enabled := false
		token := "someToken"
		wrappedHandler := NewAgent(Config{
			Token:   &token,
			Enabled: &enabled,
		}).WrapHandler(handler)

		So(wrappedHandler, ShouldEqual, handler)
	})

	Convey("Returns original handler when no token provided or is empty string", t, func() {
		handler := func() {}
		wrappedHandler := NewAgent(Config{}).WrapHandler(handler)
		emptyStringToken := ""
		wrappedHandlerEmptyStringToken := NewAgent(Config{Token: &emptyStringToken}).WrapHandler(handler)

		So(wrappedHandler, ShouldEqual, handler)
		So(wrappedHandlerEmptyStringToken, ShouldEqual, handler)
	})

	Convey("Returns wrapped handler when Config.Enabled = true", t, func() {
		handler := func() {}
		enabled := true
		token := "someToken"
		wrappedHandler := NewAgent(Config{
			Token:   &token,
			Enabled: &enabled,
		}).WrapHandler(handler)

		So(wrappedHandler, ShouldNotEqual, handler)
	})

}
