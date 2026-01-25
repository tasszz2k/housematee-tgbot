package config

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Telegram     Telegram     `mapstructure:"telegram" validate:"required"`
	GoogleApis   GoogleApis   `mapstructure:"google_apis" validate:"required"`
	GoogleSheets GoogleSheets `mapstructure:"google_sheets" validate:"required"`
}

type Telegram struct {
	ApiToken        string  `mapstructure:"api_token" validate:"required"`
	AllowedChannels []int64 `mapstructure:"allowed_channels" validate:"required"`
}

type GoogleApis struct {
	Credentials Credentials `mapstructure:"credentials" validate:"required"`
}

type Credentials struct {
	Type                string `mapstructure:"type"`
	ProjectID           string `mapstructure:"project_id"`
	PrivateKeyID        string `mapstructure:"private_key_id"`
	PrivateKey          string `mapstructure:"private_key"`
	ClientEmail         string `mapstructure:"client_email"`
	ClientID            string `mapstructure:"client_id"`
	AuthURI             string `mapstructure:"auth_uri"`
	TokenURI            string `mapstructure:"token_uri"`
	AuthProviderCertURL string `mapstructure:"auth_provider_x509_cert_url"`
	ClientCertURL       string `mapstructure:"client_x509_cert_url"`
}

type GoogleSheets struct {
	SpreadsheetId string `mapstructure:"spreadsheet_id" validate:"required"`
}

var (
	_, b, _, _        = runtime.Caller(0)
	basePath          = filepath.Dir(b) //get the absolute directory of the current file
	defaultConfigFile = basePath + "/local.yaml"
	v                 = viper.New()
	appConfig         AppConfig
)

func init() {
	Load()
}

func Load() {
	configReaderMode := os.Getenv("CONFIG_READER_MODE")

	switch configReaderMode {
	case "file":
		if err := loadConfigFromFile(); err != nil {
			panic(err)
		}
	case "secret":
		if err := loadConfigFromSecret(); err != nil {
			panic(err)
		}
	default:
		panic("Invalid CONFIG_READER_MODE. Please use 'file' or 'secret'.")
	}

	if err := scanConfigFile(&appConfig); err != nil {
		panic(err)
	}

	if err := validateConfig(&appConfig); err != nil {
		panic(err)
	}
}

func loadConfigFromFile() error {
	var configFile string
	if configFile = os.Getenv("CONFIG_PATH"); len(configFile) == 0 {
		configFile = defaultConfigFile
	}

	v.AddConfigPath(filepath.Dir(configFile))
	v.SetConfigName(
		strings.TrimSuffix(
			filepath.Base(configFile),
			filepath.Ext(configFile),
		),
	)
	v.AutomaticEnv()

	return v.ReadInConfig()
}

func loadConfigFromSecret() error {
	encodedConfig := os.Getenv("CONFIG_SECRET")
	if encodedConfig == "" {
		logrus.Fatal("CONFIG_SECRET is empty")
	}

	decodedConfig, err := base64.StdEncoding.DecodeString(encodedConfig)
	if err != nil {
		return err
	}
	v.SetConfigType("yaml")
	// Use viper to read the decoded YAML content from a buffer
	if err := v.ReadConfig(bytes.NewBuffer(decodedConfig)); err != nil {
		return err
	}

	return nil
}

func scanConfigFile(config any) error {
	return v.Unmarshal(&config)
}

func validateConfig(config any) error {
	validate := validator.New()
	return validate.Struct(config)
}

func GetAppConfig() *AppConfig {
	return &appConfig
}
