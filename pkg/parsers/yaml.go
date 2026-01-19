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

	var data map[string]any
	var pairs []*models.ConfigPair
	docCount := 0

	for {
		err := decoder.Decode(&data)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			var typeErr *yaml.TypeError
			if errors.As(err, &typeErr) {
				return nil, ErrRootNotMap
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		docCount++
		switch docCount {
		case 1:
			pairs = FlattenMap(data)
		case 2:
			fmt.Fprintf(os.Stderr, "Warning: YAML file contains multiple documents, only the first document is parsed\n")
		}

		data = nil
	}

	return pairs, nil
}
