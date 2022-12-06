package app

import (
	"encoding/json"
	"fmt"
	"likemovie/internal/search"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type App struct {
	debug  bool
	models search.Models
}

func New(debug bool) *App {
	return &App{debug: debug}
}

func (app *App) Init(searchDir string) {
	logging.SetDateRotate(logging.DateRotateConfig{
		Dir:         "./logs", // TODO
		Name:        "likemovie",
		Rotate:      7,
		WriteStdout: true,
		WriteFile:   true,
	})

	var err error
	app.models, err = search.Load(searchDir)
	runtime.Assert(err)
	nodes := app.models.Query("abc", app.debug)
	data, _ := json.MarshalIndent(nodes, "", "  ")
	fmt.Println(string(data))
}
