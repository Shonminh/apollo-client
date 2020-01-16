package agollo

import (
	"encoding/json"
	"github.com/Shonminh/apollo-client/internal/apollo"
	"github.com/Shonminh/apollo-client/internal/logger"
	"github.com/pkg/errors"
	"os"
)

type diskConfig struct {
	*apollo.Config
}

// write config to file
func writeConfigFile(config *apollo.Config, configPath string) error {
	if config == nil {
		logger.LogError("apollo config is null can not write backup file")
		return errors.New("apollo config is null can not write backup file")
	}
	file, e := os.Create(configPath)
	if e != nil {
		logger.LogError("writeConfigFile fail: %v", e)
		return e
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	dConfig := diskConfig{Config: config}
	return encoder.Encode(dConfig)
}

// load config from file
func loadConfigFile(configPath string) (*apollo.Config, error) {
	logger.LogInfo("load config file from: %v", configPath)
	file, e := os.Open(configPath)
	if e != nil {
		return nil, errors.WithMessage(e, "os.Open")
	}
	defer file.Close()

	dConfig := &diskConfig{}
	e = json.NewDecoder(file).Decode(dConfig)

	if e != nil {
		return nil, errors.WithMessage(e, "json Decode")
	}

	return dConfig.Config, nil
}
