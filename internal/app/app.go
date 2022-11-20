package app

import (
	"likemovie/internal/search"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type App struct {
	models []search.Model
}

func New() *App {
	return &App{}
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
	logging.Info("%v", app.models)
}
