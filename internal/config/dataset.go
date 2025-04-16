package config

import (
	"os"
	"path"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Dataset struct {
	Train      string   `yaml:"train" json:"train"`
	Valid      string   `yaml:"valid" json:"valid"`
	Test       string   `yaml:"test" json:"test"`
	NamesCount int      `yaml:"nc" json:"nc"`
	Names      []string `yaml:"names" json:"names"`
	namesIndex map[string]int
}

func (c Dataset) Debug() {
	log.Debug().Any("Names Index", c.namesIndex).Send()
}

func (c Dataset) GetClassName(id int) string {
	if id > len(c.Names)-1 {
		return ""
	}
	return c.Names[id]
}
func (c Dataset) GetClassId(name string) int {
	return c.namesIndex[name]
}

func (c Dataset) ToString() string {
	j, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(j)
}

func (c Dataset) YAMLMarshal() ([]byte, error) {
	return yaml.Marshal(c)
}

func NewDataset(names ...string) *Dataset {
	var namesIndex = make(map[string]int)
	for i, n := range names {
		namesIndex[n] = i
	}

	return &Dataset{
		Train:      "../train/images",
		Valid:      "../valid/images",
		Test:       "../test/images",
		Names:      names,
		NamesCount: len(names),
		namesIndex: namesIndex,
	}
}

func LoadDataset(fs afero.Fs, src string) (*Dataset, error) {
	f, err := afero.ReadFile(fs, src)
	if err != nil {
		return nil, err
	}

	conf := new(Dataset)
	err = yaml.Unmarshal(f, &conf)

	conf.namesIndex = make(map[string]int)
	for i, n := range conf.Names {
		conf.namesIndex[n] = i
	}

	return conf, err
}

func SaveDataset(fs afero.Fs, conf Dataset, dest string) error {
	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return afero.WriteFile(fs, path.Join(dest, "data.yaml"), b, os.ModePerm)
}
