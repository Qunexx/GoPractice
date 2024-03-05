package main

import (
	"Qunexx/AuthService/Configs"
	"Qunexx/AuthService/Handlers"
	"fmt"
	"net/http"
)

func main() {
	Configs.InitYamlConfig()
	Configs.InitEnvConfig()

	http.HandleFunc("/auth/login", Handlers.LoginHandler)
	http.HandleFunc("/auth/verify", Handlers.VerifyHandler)

	//Регистрация пользователя для теста
	//salt, _ := Auth.GenerateSalt(16)
	//Auth.RegisterUser("user", "user", "", salt)

	fmt.Println("Сервер запущен")
	http.ListenAndServe(Configs.EnvConfigs.ServerPort, nil)
}
