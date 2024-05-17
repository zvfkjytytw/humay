package humayapp

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func GetConfig(configFile string) (config *AppConfig, err error) {
	filename, err := filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	yamlConfig, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlConfig, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
