package Configs

import (
	"log"

	"github.com/spf13/viper"
)

// Инициализация Env файла
var EnvConfigs *envConfigs

func InitEnvConfig() {
	EnvConfigs = loadEnvVariables()
}

type envConfigs struct {
	JwtToken    string `mapstructure:"JWT_TOKEN"`
	ServerPort  string `mapstructure:"LOCAL_SERVER_PORT"`
	LoggerLevel string `mapstructure:"Logger_Level"`
}

func loadEnvVariables() (config *envConfigs) {
	viper.AddConfigPath(".")

	viper.SetConfigName("Auth")

	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading env file", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}
	return
}
