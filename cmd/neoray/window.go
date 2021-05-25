package main

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type UIOptions struct {
	arabicshape   bool
	ambiwidth     string
	emoji         bool
	guifont       string
	guifontset    string
	guifontwide   string
	linespace     int
	pumblend      int
	showtabline   int
	termguicolors bool
}

type Window struct {
	handle *sdl.Window
	title  string
}

func CreateWindow(width int, height int, title string) Window {
	window := Window{
		title: title,
	}

	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)

	sdl_window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		int32(width), int32(height), sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL)
	if err != nil {
		log_message(LOG_LEVEL_FATAL, LOG_TYPE_NEORAY, "Failed to initialize SDL window:", err)
	}
	window.handle = sdl_window

	return window
}

func (window *Window) HandleWindowResizing(editor *Editor) {
	w, h := window.handle.GetSize()
	if w != int32(GLOB_WindowWidth) || h != int32(GLOB_WindowHeight) {
		GLOB_WindowWidth = int(w)
		GLOB_WindowHeight = int(h)
		editor.nvim.ResizeUI(editor)
		editor.renderer.Resize()
	}
}

func (window *Window) Update(editor *Editor) {
	window.HandleWindowResizing(editor)
	HandleNvimRedrawEvents(editor)
	editor.cursor.Update(editor)
	// DEBUG
	fps_string := fmt.Sprintf(" | FPS: %d", GLOB_FramesPerSecond)
	idx := strings.LastIndex(window.title, " | ")
	if idx == -1 {
		window.SetTitle(window.title + fps_string)
	} else {
		window.SetTitle(window.title[0:idx] + fps_string)
	}
}

func (window *Window) SetSize(newWidth int, newHeight int, editor *Editor) {
	window.handle.SetSize(int32(newWidth), int32(newHeight))
	window.HandleWindowResizing(editor)
}

func (window *Window) SetTitle(title string) {
	window.handle.SetTitle(title)
	window.title = title
}

func (window *Window) Close() {
	window.handle.Destroy()
}
