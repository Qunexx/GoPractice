package main

import (
	"Qunexx/AuthService/Configs"
	"Qunexx/AuthService/Handlers"
	"Qunexx/AuthService/Middleware"
	"Qunexx/AuthService/logger"
	"fmt"
	"net/http"
)

func main() {
	Configs.InitYamlConfig()
	Configs.InitEnvConfig()

	logger := Logger.SetupLogger()
	mux := http.NewServeMux()

	//http.HandleFunc("/auth/login", Handlers.LoginHandler)
	//http.HandleFunc("/auth/verify", Handlers.VerifyHandler)
	mux.Handle("/auth/login", Middleware.TraceMiddleware(logger)(Middleware.RecoveryMiddleware(logger)(Middleware.LoggingMiddleware(logger)(http.HandlerFunc(Handlers.LoginHandler)))))
	mux.Handle("/auth/verify", Middleware.TraceMiddleware(logger)(Middleware.RecoveryMiddleware(logger)(Middleware.LoggingMiddleware(logger)(http.HandlerFunc(Handlers.VerifyHandler)))))

	////Регистрация пользователя для теста
	//salt, _ := Auth.GenerateSalt(16)
	//Auth.RegisterUser("user500", "user", "", salt)

	fmt.Println("Сервер запущен")
	http.ListenAndServe(Configs.EnvConfigs.ServerPort, mux)
}
