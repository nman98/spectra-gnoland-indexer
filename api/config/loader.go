package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.yaml.in/yaml/v4"
)

type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

type YamlFileReader struct{}

func (r *YamlFileReader) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

type EnvFileReader interface {
	ReadFile(name string) error
}

type DefaultEnvFileReader struct{}

func (r *DefaultEnvFileReader) ReadFile(name string) error {
	envPath := name
	fileInfo, err := os.Stat(name)
	if err == nil && fileInfo.IsDir() {
		envPath = filepath.Join(name, ".env")
	}

	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return err
	}

	// Check for file existence first
	if _, err := os.Stat(absPath); err == nil {
		// File exists, load it
		err = godotenv.Load(absPath)
		if err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}
	// If file doesn't exist, that's OK - we'll use defaults or OS env vars

	return nil
}

func LoadConfig(reader FileReader, path string) (*ApiConfig, error) {
	yamlFile, err := reader.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config ApiConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	// check if any of the fields are empty
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	// any cors method should be auto filled by the cors middleware
	return &config, nil
}

func LoadEnvironment(reader EnvFileReader, path string) (*ApiEnv, error) {
	err := reader.ReadFile(path)
	if err != nil {
		return nil, err
	}

	environment := ApiEnv{}
	if err := env.Parse(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &environment, nil
}

func LoadValkeyEnvironment(reader EnvFileReader, path string) (*ValkeyEnv, error) {
	err := reader.ReadFile(path)
	if err != nil {
		return nil, err
	}

	environment := ValkeyEnv{}
	if err := env.Parse(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &environment, nil
}
