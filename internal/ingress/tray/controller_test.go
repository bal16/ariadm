package tray_test

import (
	"testing"

	"ariadm/internal/ingress/tray"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWindow struct {
	visible bool
	shows   int
	hides   int
	focuses int
}

func (w *fakeWindow) Show() {
	w.shows++
	w.visible = true
}

func (w *fakeWindow) Hide() {
	w.hides++
	w.visible = false
}

func (w *fakeWindow) Focus()          { w.focuses++ }
func (w *fakeWindow) IsVisible() bool { return w.visible }

type fakeApplication struct {
	quits int
}

func (a *fakeApplication) Quit() { a.quits++ }

func TestMenu(t *testing.T) {
	items := tray.Menu()

	require.Len(t, items, 4)
	assert.Equal(t, tray.Item{Label: "Show AriaDM", Action: tray.ActionShow}, items[0])
	assert.Equal(t, tray.Item{Label: "Hide to Tray", Action: tray.ActionHide}, items[1])
	assert.Empty(t, items[2].Label, "third entry should be a separator")
	assert.Equal(t, tray.Item{Label: "Quit", Action: tray.ActionQuit}, items[3])
}

func TestActionString(t *testing.T) {
	tests := []struct {
		action tray.Action
		want   string
	}{
		{tray.ActionShow, "show"},
		{tray.ActionHide, "hide"},
		{tray.ActionQuit, "quit"},
		{tray.Action(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.action.String())
	}
}

func TestHandleShowShowsAndFocusesExistingWindow(t *testing.T) {
	window := &fakeWindow{}
	createCalls := 0
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return window },
		CreateWindow: func() tray.Window {
			createCalls++
			return nil
		},
	})

	controller.Handle(tray.ActionShow)

	assert.Equal(t, 1, window.shows)
	assert.Equal(t, 1, window.focuses)
	assert.Zero(t, createCalls, "existing window should not be recreated")
}

func TestHandleShowRecreatesReleasedWindow(t *testing.T) {
	var current tray.Window
	recreated := &fakeWindow{}
	createCalls := 0
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return current },
		CreateWindow: func() tray.Window {
			createCalls++
			current = recreated
			return recreated
		},
	})

	controller.Handle(tray.ActionShow)

	assert.Equal(t, 1, createCalls)
	assert.Equal(t, 1, recreated.shows)
	assert.Equal(t, 1, recreated.focuses)
}

func TestHandleShowDoesNothingWhenWindowCannotBeRecreated(t *testing.T) {
	controller := tray.NewController(tray.Config{
		Window:       func() tray.Window { return nil },
		CreateWindow: func() tray.Window { return nil },
	})

	assert.NotPanics(t, func() { controller.Handle(tray.ActionShow) })
}

func TestHandleShowWithoutWindowHooksDoesNotPanic(t *testing.T) {
	controller := tray.NewController(tray.Config{})

	assert.NotPanics(t, func() { controller.Handle(tray.ActionShow) })
}

func TestHandleHideHidesExistingWindowByDefault(t *testing.T) {
	window := &fakeWindow{visible: true}
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return window },
	})

	controller.Handle(tray.ActionHide)

	assert.Equal(t, 1, window.hides)
	assert.False(t, window.visible)
}

func TestHandleHideUsesCustomHideWindow(t *testing.T) {
	window := &fakeWindow{visible: true}
	hideCalls := 0
	controller := tray.NewController(tray.Config{
		Window:     func() tray.Window { return window },
		HideWindow: func() { hideCalls++ },
	})

	controller.Handle(tray.ActionHide)

	assert.Equal(t, 1, hideCalls)
	assert.Zero(t, window.hides, "custom hide should replace Window.Hide")
}

func TestHandleHideWithoutWindowDoesNotPanic(t *testing.T) {
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return nil },
	})

	assert.NotPanics(t, func() { controller.Handle(tray.ActionHide) })
}

func TestHandleQuitCallsApplicationQuit(t *testing.T) {
	app := &fakeApplication{}
	controller := tray.NewController(tray.Config{App: app})

	controller.Handle(tray.ActionQuit)

	assert.Equal(t, 1, app.quits)
}

func TestHandleQuitWithoutApplicationDoesNotPanic(t *testing.T) {
	controller := tray.NewController(tray.Config{})

	assert.NotPanics(t, func() { controller.Handle(tray.ActionQuit) })
}

func TestHandleIgnoresUnknownAction(t *testing.T) {
	window := &fakeWindow{visible: true}
	app := &fakeApplication{}
	controller := tray.NewController(tray.Config{
		App:    app,
		Window: func() tray.Window { return window },
	})

	controller.Handle(tray.Action(99))

	assert.Zero(t, window.shows)
	assert.Zero(t, window.hides)
	assert.Zero(t, window.focuses)
	assert.Zero(t, app.quits)
}

func TestToggleHidesVisibleWindow(t *testing.T) {
	window := &fakeWindow{visible: true}
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return window },
	})

	controller.Toggle()

	assert.Equal(t, 1, window.hides)
	assert.Zero(t, window.shows)
}

func TestToggleShowsHiddenWindow(t *testing.T) {
	window := &fakeWindow{visible: false}
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return window },
	})

	controller.Toggle()

	assert.Equal(t, 1, window.shows)
	assert.Equal(t, 1, window.focuses)
	assert.True(t, window.visible)
}

func TestToggleRecreatesMissingWindow(t *testing.T) {
	var current tray.Window
	recreated := &fakeWindow{}
	controller := tray.NewController(tray.Config{
		Window: func() tray.Window { return current },
		CreateWindow: func() tray.Window {
			current = recreated
			return recreated
		},
	})

	controller.Toggle()

	assert.Equal(t, 1, recreated.shows)
	assert.Equal(t, 1, recreated.focuses)
}

func TestToggleUsesCustomHideWindowForVisibleWindow(t *testing.T) {
	window := &fakeWindow{visible: true}
	hideCalls := 0
	controller := tray.NewController(tray.Config{
		Window:     func() tray.Window { return window },
		HideWindow: func() { hideCalls++ },
	})

	controller.Toggle()

	assert.Equal(t, 1, hideCalls)
	assert.Zero(t, window.hides)
}
