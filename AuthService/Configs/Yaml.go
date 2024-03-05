package Configs

import (
	"github.com/spf13/viper"
	"log"
)

// Инициализация yaml файла

func InitYamlConfig() {
	loadYamlPath()
	doesYamlExistCheck()
}

func loadYamlPath() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
}

func doesYamlExistCheck() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				log.Printf("Creating a new config file: %s", err)
				return
			}
		} else {
			return
		}
	}
}
