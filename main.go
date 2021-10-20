package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type state struct {
	key   string
	time  int
	label string
	next  *state
}

type NSUserNotificationCenter struct {
	objc.Object
}

type NSUserNotification struct {
	objc.Object
}

var NSUserNotification_ = objc.Get("NSUserNotification")
var NSUserNotificationCenter_ = objc.Get("NSUserNotificationCenter")

func main() {
	runtime.LockOSThread()

	s0 := state{"loaded", 0, "🕶 twentyx3 %02d:%02d", nil}
	s1 := state{"working", 1200, "🟢 Screen Time %02d:%02d", nil}
	s2 := state{"break", 20, "🟡 Break Time %02d:%02d", nil}

	s0.next = &s1
	s1.next = &s2
	s2.next = &s1

	app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle("twentyx3")

		center := NSUserNotificationCenter{NSUserNotificationCenter_.Send("defaultUserNotificationCenter")}

		nextClicked := make(chan bool)
		go func() {
			state := s0
			timer := 1200
			countdown := false
			for {
				select {
				case <-time.After(1 * time.Second):
					if timer > 0 && countdown {
						timer = timer - 1
					}
					if timer <= 0 && state != s0 {
						state = *state.next
						timer = state.time
						if state == s2 {
							notification := NSUserNotification{NSUserNotification_.Alloc().Init()}
							notification.Set("title:", core.String("twentyx3"))
							notification.Set("subtitle:", core.String("Time for a break."))
							notification.Set("informativeText:", core.String("Look at something twenty feet away for twenty seconds."))

							center.Send("deliverNotification:", notification)
							notification.Release()
						}
					}

				case <-nextClicked:
					state = *state.next
					timer = state.time
					if state == s0 {
						countdown = false
					} else {
						countdown = true
					}
				}
				// updates to the ui should happen on the main thread to avoid strange bugs
				core.Dispatch(func() {
					obj.Button().SetTitle(fmt.Sprintf(state.label, timer/60, timer%60))
				})
			}
		}()

		itemNext := cocoa.NSMenuItem_New()
		itemNext.SetTitle("Next")
		itemNext.SetAction(objc.Sel("nextClicked:"))
		cocoa.DefaultDelegateClass.AddMethod("nextClicked:", func(_ objc.Object) {
			nextClicked <- true
		})

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit")
		itemQuit.SetAction(objc.Sel("terminate:"))

		menu := cocoa.NSMenu_New()
		menu.AddItem(itemNext)
		menu.AddItem(itemQuit)
		obj.SetMenu(menu)

	})

	nsbundle := cocoa.NSBundle_Main().Class()

	nsbundle.AddMethod("__bundleIdentifier", func(_ objc.Object) objc.Object {
		return core.String("com.bradcypert.twentyx3")
	})
	nsbundle.Swizzle("bundleIdentifier", "__bundleIdentifier")

	app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyRegular)
	app.ActivateIgnoringOtherApps(true)
	app.Run()
}
