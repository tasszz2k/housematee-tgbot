package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type AppConfig struct {
	Telegram     Telegram     `mapstructure:"telegram" validate:"required"`
	GoogleApis   GoogleApis   `mapstructure:"google_apis" validate:"required"`
	GoogleSheets GoogleSheets `mapstructure:"google_sheets" validate:"required"`
}

type Telegram struct {
	ApiToken string `mapstructure:"api_token" validate:"required"`
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
	basePath          = filepath.Dir(b) //get absolute directory of the current file
	defaultConfigFile = basePath + "/local.yaml"
	v                 = viper.New()
	appConfig         AppConfig
)

func init() {
	Load()
}

func Load() {
	var configFile string
	if configFile = os.Getenv("CONFIG_PATH"); len(configFile) == 0 {
		configFile = defaultConfigFile
	}

	if err := loadConfigFile(configFile); err != nil {
		panic(err)
	}

	if err := scanConfigFile(&appConfig); err != nil {
		panic(err)
	}

	if err := validateConfig(&appConfig); err != nil {
		panic(err)
	}

}

func loadConfigFile(configFile string) error {
	configFileName := filepath.Base(configFile)
	configFilePath := filepath.Dir(configFile)

	v.AddConfigPath(configFilePath)
	v.SetConfigName(strings.TrimSuffix(configFileName, filepath.Ext(configFileName)))
	v.AutomaticEnv()

	return v.ReadInConfig()
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
