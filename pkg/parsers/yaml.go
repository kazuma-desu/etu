package parsers

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kazuma-desu/etu/pkg/models"

	"gopkg.in/yaml.v3"
)

var ErrRootNotMap = errors.New("YAML root must be a map, not an array or scalar")

type YAMLParser struct{}

func (p *YAMLParser) FormatName() string {
	return "yaml"
}

func (p *YAMLParser) Parse(path string) ([]*models.ConfigPair, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	var pairs []*models.ConfigPair
	docCount := 0

	for {
		var raw any
		err := decoder.Decode(&raw)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		docCount++
		switch docCount {
		case 1:
			data, ok := raw.(map[string]any)
			if !ok {
				return nil, ErrRootNotMap
			}
			pairs = FlattenMap(data)
		case 2:
			fmt.Fprintf(os.Stderr, "Warning: YAML file contains multiple documents, only the first document is parsed\n")
		}
	}

	return pairs, nil
}
