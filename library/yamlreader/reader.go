package yamlreader

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var PathIsDirectory = errors.New("это директория, а не yaml-файл")

// NewConfig returns the filled T structure
func NewConfig[T any](path string) (*T, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	cfg := new(T)

	if decodeErr := yaml.NewDecoder(file).Decode(&cfg); decodeErr != nil {
		return nil, decodeErr
	}

	return cfg, nil
}

func validatePath(path string) error {
	s, err := os.Stat(path)

	if err != nil {
		return err
	}

	if s.IsDir() {
		return PathIsDirectory
	}

	return nil
}
