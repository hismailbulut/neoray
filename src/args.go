package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/sqweek/dialog"
)

var usageTemplate = `
Neoray is an ui client for neovim.
Copyright (c) 2021 Ismail Bulut.
Version %d.%d.%d %s
License %s
Webpage %s

Options:

--file <name>
	Filename to open.
--line <number>
	Cursor goes to line <number>.
--column <number>
	Cursor goes to column <number>.
--singleinstance, -si
	Only accepts one instance of neoray and sends all flags to it.
--verbose
	Prints verbose debug output to a file.
--nvim <path>
	Path to nvim executable. May be relative or absolute.
--server <address>
	Connect to existing neovim instance.
--multigrid
	Enables multigrid support.
--version, -v
	Prints only the version and quits.
--help, -h
	Prints this message and quits.

All other flags will send to neovim.
`

type ParsedArgs struct {
	file       string
	line       int
	column     int
	singleInst bool
	execPath   string
	address    string
	multiGrid  bool
	others     []string
}

func ParseArgs(args []string) ParsedArgs {
	// Init defaults
	options := ParsedArgs{
		file:       "",
		line:       -1,
		column:     -1,
		singleInst: false,
		execPath:   "nvim",
		address:    "",
		others:     []string{},
	}
	var err error
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			assert(len(args) > i+1, "specify filename after --file")
			options.file = args[i+1]
			i++
			break
		case "--line":
			assert(len(args) > i+1, "specify line number after --line")
			options.line, err = strconv.Atoi(args[i+1])
			assert(err == nil, "invalid number after --line")
			i++
			break
		case "--column":
			assert(len(args) > i+1, "specify column number after --column")
			options.column, err = strconv.Atoi(args[i+1])
			assert(err == nil, "invalid number after --column")
			i++
			break
		case "--singleinstance", "-si":
			options.singleInst = true
			break
		case "--verbose":
			initVerboseFile("neoray_verbose.log")
			break
		case "--nvim":
			assert(len(args) > i+1, "specify path after --nvim")
			absolute, err := filepath.Abs(args[i+1])
			if err == nil {
				options.execPath = absolute
			}
			i++
			break
		case "--server":
			assert(len(args) > i+1, "specify address after --server")
			options.address = args[i+1]
			i++
			break
		case "--multigrid":
			options.multiGrid = true
		case "--version", "-v":
			PrintVersion()
			os.Exit(0)
		case "--help", "-h":
			PrintHelp()
			os.Exit(0)
			break
		default:
			options.others = append(options.others, args[i])
			break
		}
	}
	return options
}

func PrintVersion() {
	msg := "Neoray " + versionString() + "\n" + "Start with -h option for more information."
	switch runtime.GOOS {
	case "windows":
		dialog.Message(msg).Title("Version").Info()
	default:
		fmt.Println(msg)
	}
}

func PrintHelp() {
	// About
	msg := fmt.Sprintf(usageTemplate,
		VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH,
		buildTypeString(), LICENSE, WEBPAGE)
	dialog.Message(msg).Title("Help").Info()
	if runtime.GOOS != "windows" {
		// Also print help to stdout for linux and darwin
		fmt.Println(msg)
	}
}

// Call this before starting neovim.
func (options ParsedArgs) ProcessBefore() bool {
	if options.singleInst {
		// First we will check only once because sending and
		// waiting http requests will make neoray opens slower.
		client, err := CreateClient()
		if err != nil {
			logMessage(LEVEL_DEBUG, TYPE_NETWORK, "No instance found or ipc client creation failed:", err)
			return false
		}
		defer client.Close()
		if options.file != "" {
			fullPath, err := filepath.Abs(options.file)
			if err == nil {
				if !client.Call(IPC_MSG_TYPE_OPEN_FILE, fullPath) {
					return false
				}
			}
		}
		if options.line != -1 {
			if !client.Call(IPC_MSG_TYPE_GOTO_LINE, options.line) {
				return false
			}
		}
		if options.column != -1 {
			if !client.Call(IPC_MSG_TYPE_GOTO_COLUMN, options.column) {
				return false
			}
		}
		return true
	}
	return false
}

// Call this after connected neovim as ui.
func (options ParsedArgs) ProcessAfter() {
	if options.singleInst {
		server, err := CreateServer()
		if err != nil {
			logMessage(LEVEL_ERROR, TYPE_NEORAY, "Failed to create ipc server:", err)
		} else {
			singleton.server = server
			logMessage(LEVEL_TRACE, TYPE_NEORAY, "Ipc server created.")
		}
	}
	if options.file != "" {
		singleton.nvim.openFile(options.file)
	}
	if options.line != -1 {
		singleton.nvim.gotoLine(options.line)
	}
	if options.column != -1 {
		singleton.nvim.gotoColumn(options.column)
	}
}
