package search

import "fmt"

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
			Uploader Field `yaml:"uploader"` // uploader
			Detail   Field `yaml:"detail"`   // detail url
		} `yaml:"fields"`
	} `yaml:"parser"`
}

type Field struct {
	Selector   string `yaml:"selector"`
	Attribute  string `yaml:"attribute"`
	TimeFormat string `yaml:"time_format"`
	Request    *Field `yaml:"request"`
}

func (m Model) String() string {
	return fmt.Sprintf("%s => %s", m.Name, m.Search)
}
