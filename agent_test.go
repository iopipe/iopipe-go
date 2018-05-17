package iopipe

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	Convey("An agent created with empty Config should have default configuration", t, func() {
		agentWithDefaultConfig := NewAgent(AgentConfig{})

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
			inputConfig := AgentConfig{
				Token:         test.inputToken,
				Enabled:       &test.inputEnabled,
				TimeoutWindow: &test.inputTimeoutWindow,
			}

			agentWithCustomConfig := NewAgent(inputConfig)

			Convey(fmt.Sprintf(`Config with inputToken "%s" should be mapped corectly in the agent instance`, test.inputToken), func() {
				So(agentWithCustomConfig.Token, ShouldEqual, test.inputToken)
				So(*agentWithCustomConfig.Enabled, ShouldEqual, test.inputEnabled)
				So(*agentWithCustomConfig.TimeoutWindow, ShouldEqual, test.inputTimeoutWindow)
				So(agentWithCustomConfig.Plugins, ShouldBeEmpty)
			})
		}

	})
}

func TestWrapHandler(t *testing.T) {
	Convey("Returns original handler when Config.Enabled = false", t, func() {
		handler := func() {}
		enabled := false
		wrappedHandler := NewAgent(AgentConfig{
			Enabled: &enabled,
		}).WrapHandler(handler)

		So(wrappedHandler, ShouldEqual, handler)
	})

	Convey("Returns wrapped handler when Config.Enabled = true", t, func() {
		handler := func() {}
		enabled := true
		wrappedHandler := NewAgent(AgentConfig{
			Enabled: &enabled,
		}).WrapHandler(handler)

		So(wrappedHandler, ShouldNotEqual, handler)
	})
}
