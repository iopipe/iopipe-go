package iopipe

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAgent_NewAgent(t *testing.T) {
	Convey("An agent created with empty Config should have default configuration", t, func() {
		agentWithDefaultConfig := NewAgent(Config{})

		Convey("Agent should disable debug mode by default", func() {
			So(*agentWithDefaultConfig.Debug, ShouldBeFalse)
		})

		Convey("Agent should be enabled by default", func() {
			So(*agentWithDefaultConfig.Enabled, ShouldBeTrue)
		})

		Convey("Plugins should be empty by default", func() {
			So(agentWithDefaultConfig.plugins, ShouldBeEmpty)
		})

		Convey("TimeoutWindow should be 150 ms by default", func() {
			So(*agentWithDefaultConfig.TimeoutWindow, ShouldEqual, 150*time.Millisecond)
		})

		Convey("Token should be empty by default", func() {
			So(*agentWithDefaultConfig.Token, ShouldBeEmpty)
		})
	})

	Convey("An agent created with custom Config should assume that configuration", t, func() {
		var tests = []struct {
			inputEnabled       bool
			inputTimeoutWindow time.Duration
			inputToken         string
		}{
			{
				inputEnabled:       false,
				inputTimeoutWindow: 0 * time.Millisecond,
				inputToken:         "",
			},
			{
				inputEnabled:       true,
				inputTimeoutWindow: 20 * time.Millisecond,
				inputToken:         "non-empty-value",
			},
		}

		for _, test := range tests {
			inputConfig := Config{
				Enabled:       &test.inputEnabled,
				TimeoutWindow: &test.inputTimeoutWindow,
				Token:         &test.inputToken,
			}

			agentWithCustomConfig := NewAgent(inputConfig)

			Convey(fmt.Sprintf(`Config with inputToken "%s" should be mapped corectly in the agent instance`, test.inputToken), func() {
				So(*agentWithCustomConfig.Enabled, ShouldEqual, test.inputEnabled)
				So(*agentWithCustomConfig.TimeoutWindow, ShouldEqual, test.inputTimeoutWindow)
				So(*agentWithCustomConfig.Token, ShouldEqual, test.inputToken)
				So(agentWithCustomConfig.plugins, ShouldBeEmpty)
			})
		}
	})

	Convey("An agent should check environment variables for configuration", t, func() {
		Convey("IOPIPE_DEBUG should enable debug mode", func() {
			oldValue := os.Getenv("IOPIPE_DEBUG")
			os.Setenv("IOPIPE_DEBUG", "true")

			a := NewAgent(Config{})
			So(*a.Debug, ShouldBeTrue)

			os.Setenv("IOPIPE_DEBUG", oldValue)
		})

		Convey("IOPIPE_ENABLED should disable agent", func() {
			oldValue := os.Getenv("IOPIPE_ENABLED")
			os.Setenv("IOPIPE_DEBUG", "false")

			a := NewAgent(Config{})
			So(*a.Debug, ShouldBeFalse)

			os.Setenv("IOPIPE_DEBUG", oldValue)
		})

		Convey("IOPIPE_TIMEOUT_WINDOW should set the timeout window", func() {
			oldValue := os.Getenv("IOPIPE_TIMEOUT_WINDOW")
			os.Setenv("IOPIPE_TIMEOUT_WINDOW", "300")

			a := NewAgent(Config{})
			So(*a.TimeoutWindow, ShouldEqual, time.Duration(300*time.Millisecond))

			os.Setenv("IOPIPE_DEBUG", oldValue)
		})
	})

	Convey("An agent with plugins should initialize them", t, func() {
		agentWithPlugin := NewAgent(Config{
			PluginInstantiators: []PluginInstantiator{
				TestPlugin(TestPluginConfig{}),
			},
		})

		Convey("Should only have one plugin", func() {
			So(len(agentWithPlugin.plugins), ShouldEqual, 1)
		})

		plugin, _ := agentWithPlugin.plugins[0].(*testPlugin)

		Convey("Pre setup hooks should be called once", func() {
			So(plugin.preSetupCalled, ShouldEqual, 1)
		})

		Convey("Post setup hooks should be called once", func() {
			So(plugin.postSetupCalled, ShouldEqual, 1)
		})

		Convey("No other rhooks should be called", func() {
			So(plugin.preInvokeCalled, ShouldEqual, 0)
			So(plugin.postInvokeCalled, ShouldEqual, 0)
			So(plugin.preReportCalled, ShouldEqual, 0)
			So(plugin.postReportCalled, ShouldEqual, 0)
		})
	})
}

func TestAgent_WrapHandler(t *testing.T) {
	Convey("Returns original handler when Config.Enabled = false", t, func() {
		handler := func() {}
		enabled := false
		token := "someToken"
		wrappedHandler := NewAgent(Config{
			Enabled: &enabled,
			Token:   &token,
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
