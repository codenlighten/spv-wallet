package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/BuxOrg/bux-server/dictionary"
)

// Added a mutex lock for a race-condition
var viperLock sync.Mutex

// Load all AppConfig
func Load() (appConfig *AppConfig, err error) {
	viperLock.Lock()
	defer viperLock.Unlock()

	if err = setDefaults(); err != nil {
		return nil, err
	}

	envConfig()

	if err = loadFlags(); err != nil {
		return nil, err
	}

	if err = loadFromFile(); err != nil {
		return nil, err
	}

	appConfig = new(AppConfig)
	if err = unmarshallToAppConfig(appConfig); err != nil {
		return nil, err
	}

	return appConfig, nil
}

func setDefaults() error {
	viper.SetDefault(ConfigFilePathKey, DefaultConfigFilePath)

	defaultsMap := make(map[string]interface{})
	if err := mapstructure.Decode(DefaultAppConfig, &defaultsMap); err != nil {
		return err
	}

	for key, value := range defaultsMap {
		viper.SetDefault(key, value)
	}

	return nil
}

func envConfig() {
	viper.SetEnvPrefix("BUX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func loadFromFile() error {
	configFilePath := viper.GetString(ConfigFilePathKey)

	if configFilePath == DefaultConfigFilePath {
		_, err := os.Stat(DefaultConfigFilePath)
		if os.IsNotExist(err) {
			// if the config is not specified and no default config file exists, use defaults
			logger.Data(2, logger.DEBUG, "Config file not specified. Using defaults")
			return nil
		}
		configFilePath = DefaultConfigFilePath
	}

	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf(dictionary.GetInternalMessage(dictionary.ErrorReadingConfig), err.Error())
		logger.Data(2, logger.ERROR, err.Error())
		return err
	}

	return nil
}

func unmarshallToAppConfig(appConfig *AppConfig) error {
	if err := viper.Unmarshal(&appConfig); err != nil {
		err = fmt.Errorf(dictionary.GetInternalMessage(dictionary.ErrorViper), err.Error())
		return err
	}
	return nil
}
