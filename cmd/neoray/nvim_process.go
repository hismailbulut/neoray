package main

import (
	"os"
	"sync"

	"github.com/neovim/go-client/nvim"
)

type NvimProcess struct {
	handle       *nvim.Nvim
	update_mutex *sync.Mutex
	update_stack [][][]interface{}
}

func CreateNvimProcess() NvimProcess {
	proc := NvimProcess{
		update_mutex: &sync.Mutex{},
		update_stack: make([][][]interface{}, 0),
	}

	// TODO: preprocess args in main function
	args := []string{
		"--embed",
		// "-u",
		// "NORC",
		// "--noplugin",
	}
	args = append(args, os.Args[1:]...)

	nv, err := nvim.NewChildProcess(
		nvim.ChildProcessServe(true),
		nvim.ChildProcessArgs(args...))
	if err != nil {
		log_message(LOG_LEVEL_FATAL, LOG_TYPE_NVIM, err)
	}

	proc.handle = nv
	proc.requestApiInfo()
	proc.introduce()

	// proc.ExecuteVimScript("au UIEnter call rpcnotify(0, \"uienter_success\")")
	// proc.handle.RegisterHandler("uienter_success", func() {
	//     log_debug_msg("UI Entered Succesfully.")
	// })

	log_message(LOG_LEVEL_DEBUG, LOG_TYPE_NVIM, "Neovim child process created.")

	return proc
}

func (proc *NvimProcess) requestApiInfo() {
	_, err := proc.handle.APIInfo()
	if err != nil {
		log_message(LOG_LEVEL_ERROR, LOG_TYPE_NVIM, "Failed to get api information:", err)
		return
	}
	// for _, info := range apiInfo[1:] {
	//     mapIter := reflect.ValueOf(info).MapRange()
	//     for mapIter.Next() {
	//         fmt.Println("Key:", mapIter.Key(), "Val:", mapIter.Value())
	//     }
	// }
}

func (proc *NvimProcess) introduce() {
	// Short name for the connected client
	name := NEORAY_NAME
	// Dictionary describing the version
	version := &nvim.ClientVersion{
		Major:      NEORAY_VERSION_MAJOR,
		Minor:      NEORAY_VERSION_MINOR,
		Patch:      NEORAY_VERSION_PATCH,
		Prerelease: "dev",
		Commit:     "main",
	}
	// Client type
	typ := "ui"
	// Builtin methods in the client
	methods := make(map[string]*nvim.ClientMethod, 0)
	// Arbitrary string:string map of informal client properties
	attributes := make(nvim.ClientAttributes, 1)
	attributes["website"] = NEORAY_WEBPAGE
	attributes["license"] = NEORAY_LICENSE

	err := proc.handle.SetClientInfo(name, version, typ, methods, attributes)
	if err != nil {
		log_message(LOG_LEVEL_ERROR, LOG_TYPE_NVIM, "Failed to set client information:", err)
	}
}

func (proc *NvimProcess) ExecuteVimScript(script string) {
	if err := proc.handle.Command(script); err != nil {
		log_message(LOG_LEVEL_ERROR, LOG_TYPE_NVIM, "Failed to execute vimscript:", err)
	}
}

func (proc *NvimProcess) SendKeyCode(keycode string) {
	written, err := proc.handle.Input(keycode)
	if err != nil {
		log_message(LOG_LEVEL_WARN, LOG_TYPE_NVIM, "Failed to send input keys:", err)
	}
	if written != len(keycode) {
		log_message(LOG_LEVEL_WARN, LOG_TYPE_NVIM, "Failed to send some keys.")
	}
}

func (proc *NvimProcess) SendButton(button, action, modifier string, grid, row, column int) {
	err := proc.handle.InputMouse(button, action, modifier, grid, row, column)
	if err != nil {
		log_message(LOG_LEVEL_WARN, LOG_TYPE_NVIM, "Failed to send mouse input:", err)
	}
}

func (proc *NvimProcess) StartUI() {
	options := make(map[string]interface{})
	options["rgb"] = true
	options["ext_linegrid"] = true

	proc.handle.AttachUI(EditorSingleton.columnCount, EditorSingleton.rowCount, options)

	proc.handle.RegisterHandler("redraw",
		func(updates ...[]interface{}) {
			proc.update_mutex.Lock()
			proc.update_stack = append(proc.update_stack, updates)
			proc.update_mutex.Unlock()
		})

	go func() {
		if err := proc.handle.Serve(); err != nil {
			log_message(LOG_LEVEL_ERROR, LOG_TYPE_NVIM, "Neovim child process exited with errors:", err)
			return
		}
		log_message(LOG_LEVEL_DEBUG, LOG_TYPE_NVIM, "Neovim child process closed.")
		EditorSingleton.quit_requested_chan <- true
	}()

	log_message(LOG_LEVEL_DEBUG, LOG_TYPE_NVIM, "UI Connected. Rows:", EditorSingleton.rowCount, "Columns:", EditorSingleton.columnCount)
}

func (proc *NvimProcess) ResizeUI() {
	EditorSingleton.columnCount = EditorSingleton.window.width / EditorSingleton.cellWidth
	EditorSingleton.rowCount = EditorSingleton.window.height / EditorSingleton.cellHeight
	proc.handle.TryResizeUI(EditorSingleton.columnCount, EditorSingleton.rowCount)
	log_debug_msg("UI Resized. Rows:", EditorSingleton.rowCount, "Columns:", EditorSingleton.columnCount)
}

func (proc *NvimProcess) Close() {
	proc.handle.Close()
}
