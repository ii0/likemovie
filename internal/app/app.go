package app

import (
	"encoding/json"
	"fmt"
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
	for _, model := range app.models {
		nodes, err := model.Query("abc")
		if err != nil {
			logging.Error("%s: %v", model.Name, err)
			continue
		}
		data, _ := json.MarshalIndent(nodes, "", "  ")
		fmt.Println(string(data))
	}
}
