package search

import (
	"errors"
	"fmt"
)

var errNotfound = errors.New("not found")

type Model struct {
	Name   string `yaml:"name"`   // site name
	Search string `yaml:"search"` // search url
	Parser struct {
		List   string `yaml:"list"` // list selector
		Fields struct {
			ID       Field `yaml:"id"`       // id
			Name     Field `yaml:"name"`     // name
			Type     Field `yaml:"type"`     // movie/tv
			Upload   Field `yaml:"upload"`   // upload time
			Download Field `yaml:"download"` // download url
			Size     Field `yaml:"size"`     // size
			Seeders  Field `yaml:"seeders"`  // seeders
			Leechers Field `yaml:"leechers"` // leechers
			Peers    Field `yaml:"peers"`    // peers
			Uploader Field `yaml:"uploader"` // uploader
			Detail   Field `yaml:"detail"`   // detail url
		} `yaml:"fields"`
	} `yaml:"parser"`
}

type MapRule struct {
	Contain string `yaml:"contain"`
	To      string `yaml:"to"`
}

type Field struct {
	Selector    string    `yaml:"selector"`
	Attribute   string    `yaml:"attribute"`
	TimeFormats []string  `yaml:"time_formats"`
	Request     *Field    `yaml:"request"`
	Maps        []MapRule `yaml:"maps"`
}

func (m *Model) String() string {
	return fmt.Sprintf("%s => %s", m.Name, m.Search)
}
