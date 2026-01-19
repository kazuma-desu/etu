package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/models"
)

var ErrRootNotObject = errors.New("JSON root must be an object, not an array or scalar")

type JSONParser struct{}

func (p *JSONParser) FormatName() string {
	return "json"
}

func (p *JSONParser) Parse(path string) ([]*models.ConfigPair, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	rootMap, ok := root.(map[string]any)
	if !ok {
		return nil, ErrRootNotObject
	}

	return FlattenMap(rootMap), nil
}
