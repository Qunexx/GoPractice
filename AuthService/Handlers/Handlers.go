package Handlers

import (
	"Qunexx/AuthService/Auth"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Контроллер запроса авторизации
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Только POST метод разрешён", http.StatusMethodNotAllowed)
		return
	}
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	authenticated, err := Auth.Authenticate(username, password)
	if err != nil {
		log.Fatalf("Ошибка функции аутентификации %v", err)
	}
	if authenticated {
		fmt.Println("Аутентификация успешна")
		err := Auth.SetTokenCookies(w, username)
		if err != nil {
			http.Error(w, "Ошибка установления куки токенов", http.StatusInternalServerError)
			return
		} else {
			fmt.Println("Куки токенов установлены успешно")
			return
		}
	} else {
		authenticated = false
		fmt.Println("Ошибка аутентификации")
		http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
	}

}

// Контроллер запроса верификации
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Только POST метод разрешён", http.StatusMethodNotAllowed)
		return
	}

	accessTokenCookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, "Акссес токен недействителен", http.StatusUnauthorized)
		return
	}
	fmt.Println("Проверка рефреш токена")
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Рефреш токен недействителен", http.StatusUnauthorized)
		return
	} else {
		valid, err := Auth.VerifyToken(accessTokenCookie.Value)
		if valid && err == nil {
			fmt.Println("Акссестокен есть")
			expired, err := Auth.IsTokenExpired(accessTokenCookie.Value)
			if !expired && err == nil {
				fmt.Println("Статус ОК")
				userInfo, err := Auth.GetUserInfo(accessTokenCookie.Value)
				if err != nil {
					http.Error(w, "Не удалось получить информацию о пользователе", http.StatusInternalServerError)
					return
				}

				responseData, err := json.Marshal(userInfo)
				w.Write(responseData)
			} else {
				valid, _ := Auth.IsTokenExpired(refreshTokenCookie.Value)
				if valid == true {
					newAccessTokenStr, newRefreshTokenStr, err := Auth.RegenerateTokens(refreshTokenCookie.Value)
					if err == nil {
						fmt.Println("Токены перегенерированы и присвоены")
						err := Auth.SetNewTokensCookies(w, newAccessTokenStr, newRefreshTokenStr)
						if err == nil {
							w.WriteHeader(http.StatusOK)
						}
					}
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
			}
		}
	}
}
