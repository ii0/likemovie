package search

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(dir string) ([]Model, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var ret []Model
	for _, file := range files {
		model, err := load(file)
		if err != nil {
			return nil, err
		}
		ret = append(ret, model)
	}
	return ret, nil
}

func load(dir string) (Model, error) {
	var model Model
	f, err := os.Open(dir)
	if err != nil {
		return model, err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&model)
	if err != nil {
		return model, err
	}
	return model, nil
}
