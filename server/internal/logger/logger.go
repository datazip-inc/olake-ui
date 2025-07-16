package logger

import (
	"os"
	"path"
	"sync"

	"github.com/beego/beego/v2/core/logs"
)

var (
	loggerInitOnce sync.Once
)

func InitLogger(logdir string) {
	loggerInitOnce.Do(func() {
		// Clear existing loggers first
		logs.Reset()

		// Create logs directory
		if err := os.MkdirAll(logdir, 0755); err != nil {
			panic("Failed to create log directory: " + err.Error())
		}

		// Console configuration
		consoleConfig := `{
			"level": 7,
			"color": true
		}`

		// File configuration
		fileConfig := `{
			"filename": "` + path.Join(logdir, "olake-server.log") + `",
			"level": 7,
			"maxlines": 1000,
			"maxdays": 7,
			"daily": false,
			"rotate": true,
			"perm": "0644"
		}`

		// Initialize loggers
		if err := logs.SetLogger(logs.AdapterConsole, consoleConfig); err != nil {
			panic("Console logger init failed: " + err.Error())
		}

		if err := logs.SetLogger(logs.AdapterFile, fileConfig); err != nil {
			panic("File logger init failed: " + err.Error())
		}

		// Configure logger behavior
		logs.SetLogFuncCallDepth(3)
		logs.EnableFuncCallDepth(true)
		logs.SetLevel(logs.LevelDebug)
	})
}
