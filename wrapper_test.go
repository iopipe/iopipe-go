package iopipe

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

// TODO: this is where the fun begins ._.

func TestRunHook(t *testing.T) {
	Convey("RunHook should run hooks on all the plugins", t, func() {
		plugin1 := &testPlugin{}
		plugin2 := &testPlugin{}

		w := wrapper{
			plugins: []Plugin{
				plugin1,
				plugin2,
			},
		}

		Convey("All plugin should receive RunHook", func() {
			w.RunHook("test-hook")

			So(plugin1.LastHook, ShouldEqual, "test-hook")
			So(plugin2.LastHook, ShouldEqual, "test-hook")
		})

		Convey("nil plugin does not result in panic", func() {
			w.plugins = append(w.plugins, nil)

			So(func() {
				w.RunHook("not-test-hook")
			}, ShouldNotPanic)
			So(plugin1.LastHook, ShouldEqual, "not-test-hook")
			So(plugin2.LastHook, ShouldEqual, "not-test-hook")
		})


	})
}
