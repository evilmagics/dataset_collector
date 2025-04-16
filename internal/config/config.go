package config

import (
	"os"
	"path"

	// "github.com/goccy/go-yaml"

	"github.com/evilmagics/dataset_collector/internal/utils"
	"github.com/goccy/go-json"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Source struct {
	Src           string              `yaml:"src" json:"src"`
	ClassSync     utils.ClassNameSync `yaml:"class_name_sync" json:"class_name_sync"`
	DatasetConfig *Dataset            `yaml:"-" json:"data_config"`
}

func (s *Source) LoadDatasetConfig(fs afero.Fs) (err error) {
	s.DatasetConfig, err = LoadDataset(fs, path.Join(s.Src, "data.yaml"))
	if err != nil {
		return err
	}
	return err
}

type Config struct {
	Dest    string   `yaml:"dest" json:"dest"`
	Classes []string `yaml:"classes" json:"classes"`
	Sources []Source `yaml:"sources" json:"sources"`
	Workers int      `yaml:"workers" json:"workers"`
}

func (c Config) String() string {
	j, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(j)
}

func LoadConfig() (*Config, error) {
	f, err := os.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}

	conf := new(Config)
	if err = yaml.Unmarshal(f, &conf); err != nil {
		return nil, err
	}

	return conf, err
}
