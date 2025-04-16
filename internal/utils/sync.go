package utils

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type ClassNameSync struct {
	index map[string]string
	raw   map[string][]string
}

func (s ClassNameSync) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"index": s.index,
		"raw":   s.raw,
	})
}

func (s *ClassNameSync) UnmarshalYAML(node *yaml.Node) error {
	var raw map[string][]string
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// save raw classes name synchronizing
	s.raw = raw

	// inverse raw to faster indexing
	// {"car": ["car", "van"]} -> {"car": "car", "van": "car"}
	s.index = make(map[string]string)
	for k, v := range s.raw {
		for _, i := range v {
			s.index[i] = k
		}
	}

	return nil
}

func (s ClassNameSync) GetCrossName(class string) *string {
	c := s.index[class]
	if c == "" {
		return nil
	}
	return &c
}
