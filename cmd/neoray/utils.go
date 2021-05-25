package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	COLOR_WHITE       = sdl.Color{R: 255, G: 255, B: 255, A: 255}
	COLOR_BLACK       = sdl.Color{R: 0, G: 0, B: 0, A: 255}
	COLOR_TRANSPARENT = sdl.Color{R: 0, G: 0, B: 0, A: 0}
)

func convert_rgb24_to_rgba(color uint32) sdl.Color {
	return sdl.Color{
		// 0x000000rr & 0xff = red: 0xrr
		R: uint8((color >> 16) & 0xff),
		// 0x0000rrgg & 0xff = green: 0xgg
		G: uint8((color >> 8) & 0xff),
		// 0x00rrggbb & 0xff = blue: 0xbb
		B: uint8(color & 0xff),
		A: 255,
	}
}

func convert_rgba_to_rgb24(color sdl.Color) uint32 {
	// 0x00000000
	rgb24 := uint32(color.R)
	// 0x000000rr
	rgb24 = (rgb24 << 8) | uint32(color.G)
	// 0x0000rrgg
	rgb24 = (rgb24 << 8) | uint32(color.B)
	// 0x00rrggbb
	return rgb24
}

type f32color struct {
	R float32
	G float32
	B float32
	A float32
}

func u8color_to_fcolor(sdlcolor sdl.Color) f32color {
	return f32color{
		R: float32(sdlcolor.R) / 256,
		G: float32(sdlcolor.G) / 256,
		B: float32(sdlcolor.B) / 256,
		A: float32(sdlcolor.A) / 256,
	}
}

func is_color_black(color sdl.Color) bool {
	return color.R == 0 && color.G == 0 && color.B == 0
}

// Math
type f32vec2 struct {
	X, Y float32
}
type i32vec2 struct {
	X, Y int32
}
type ivec2 struct {
	X, Y int
}

func lerpf32(v0, v1, t float32) float32 {
	return v0 + t*(v1-v0)
}

func iabs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func triangulate_rect(rect *sdl.Rect) [6]i32vec2 {
	return [6]i32vec2{
		{rect.X, rect.Y},                   //0
		{rect.X, rect.Y + rect.H},          //1
		{rect.X + rect.W, rect.Y + rect.H}, //2
		{rect.X + rect.W, rect.Y + rect.H}, //2
		{rect.X + rect.W, rect.Y},          //3
		{rect.X, rect.Y},                   //0
	}
}

func triangulate_frect(rect *sdl.FRect) [6]f32vec2 {
	return [6]f32vec2{
		{rect.X, rect.Y},                   //0
		{rect.X, rect.Y + rect.H},          //1
		{rect.X + rect.W, rect.Y + rect.H}, //2
		{rect.X + rect.W, rect.Y + rect.H}, //2
		{rect.X + rect.W, rect.Y},          //3
		{rect.X, rect.Y},                   //0
	}
}

// Function execution time mesurement functions
type FunctionMeasure struct {
	totalCall int64
	totalTime time.Duration
}

var measure_averages map[string]FunctionMeasure
var measure_averages_mutex sync.Mutex

func init_function_time_tracker() {
	measure_averages = make(map[string]FunctionMeasure)
}

func measure_execution_time(name string) func() {
	now := time.Now()
	return func() {
		elapsed := time.Since(now)
		measure_averages_mutex.Lock()
		defer measure_averages_mutex.Unlock()
		if val, ok := measure_averages[name]; ok == true {
			val.totalCall++
			val.totalTime += elapsed
			measure_averages[name] = val
		} else {
			measure_averages[name] = FunctionMeasure{
				totalCall: 1,
				totalTime: elapsed,
			}
		}
	}
}

func close_function_time_tracker() {
	for key, val := range measure_averages {
		log_message(LOG_LEVEL_DEBUG, LOG_TYPE_PERFORMANCE,
			key, "Calls:", val.totalCall, "Time:", val.totalTime, "Average:", val.totalTime/time.Duration(val.totalCall))
	}
}

// Other utility functions
func has_flag_u16(val, flag uint16) bool {
	return val&flag != 0
}

func atomic_copy_f32(dst, src *float32) {
}

// Logger
const MINIMUM_LOG_LEVEL = LOG_LEVEL_DEBUG
const (
	// log levels
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
	// log types
	LOG_TYPE_NVIM
	LOG_TYPE_NEORAY
	LOG_TYPE_RENDERER
	LOG_TYPE_PERFORMANCE
	LOG_TYPE_DEBUG_MESSAGE
)

func log_message(log_level, log_type int, message ...interface{}) {
	if log_level < MINIMUM_LOG_LEVEL {
		return
	}
	log_string := " "
	debug_type := false
	switch log_type {
	case LOG_TYPE_NVIM:
		log_string += "[NVIM]"
	case LOG_TYPE_NEORAY:
		log_string += "[NEORAY]"
	case LOG_TYPE_RENDERER:
		log_string += "[RENDERER]"
	case LOG_TYPE_PERFORMANCE:
		log_string += "[PERFORMANCE]"
	case LOG_TYPE_DEBUG_MESSAGE:
		log_string += ">>"
		debug_type = true
	default:
		return
	}
	err := false
	fatal := false
	log_string += " "
	if !debug_type {
		switch log_level {
		case LOG_LEVEL_DEBUG:
			log_string += "DEBUG:"
		case LOG_LEVEL_WARN:
			log_string += "WARNING:"
		case LOG_LEVEL_ERROR:
			log_string += "ERROR:"
			err = true
		case LOG_LEVEL_FATAL:
			log_string += "FATAL:"
			fatal = true
		default:
			return
		}
	}
	log_string += " "
	for _, msg := range message {
		log_string += fmt.Sprint(msg)
		log_string += " "
	}
	if fatal {
		fmt.Printf("\n")
		debug.PrintStack()
		log.Fatalln(log_string)
	} else if err {
		log.Println(log_string)
	} else {
		log.Println(log_string)
	}
}

func log_debug_msg(message ...interface{}) {
	log_message(LOG_LEVEL_DEBUG, LOG_TYPE_DEBUG_MESSAGE, message...)
}

func assert(cond bool, message ...interface{}) {
	if cond == false {
		log_message(LOG_LEVEL_FATAL, LOG_TYPE_NEORAY, "Assertion Failed: ", message)
	}
}
