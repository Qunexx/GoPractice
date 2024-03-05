package Auth

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"

	"Qunexx/AuthService/Configs"
)

func encryptPassword(password string, salt []byte) string {
	dk := pbkdf2.Key([]byte(password), salt, 4096, 64, sha512.New)
	return hex.EncodeToString(dk)
}

func GenerateSalt(size int) ([]byte, error) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func RegisterUser(username, password, email string, salt []byte) error {
	Configs.InitYamlConfig()

	encryptedPassword := encryptPassword(password, salt)
	var users []map[string]string
	if err := viper.UnmarshalKey("users", &users); err != nil {
		return err
	}

	user := map[string]string{
		"login":    username,
		"email":    email,
		"password": encryptedPassword,
		"salt":     hex.EncodeToString(salt),
	}

	users = append(users, user)
	viper.Set("users", users)

	return viper.WriteConfigAs("config.yaml")
}

// Аутентификация пользователя
func Authenticate(login, password string) (bool, error) {
	Configs.InitYamlConfig()
	var users []map[string]string
	if err := viper.UnmarshalKey("users", &users); err != nil {
		fmt.Println("Не удалось воспользоваться информацией о пользователях:")
		return false, fmt.Errorf("Не удалось воспользоваться информацией о пользователях: %v", err)
	}

	for _, user := range users {
		if user["login"] == login {
			salt, err := hex.DecodeString(user["salt"])
			if err != nil {
				fmt.Println("Не удалось воспользоваться информацией о пользователях:")
				return false, fmt.Errorf("Ошибка дешифрации соли для пользователя %s: %v", login, err)
			}

			encryptedPassword := encryptPassword(password, salt)
			if user["password"] == encryptedPassword {
				return true, nil
			}
			break
		}
	}

	return false, nil
}

// Генерация токенов
func GenerateTokens(login string) (accessTokenStr, refreshTokenStr string, err error) {
	Key := Configs.EnvConfigs.JwtToken
	// Генерирую акссес токен
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(1 * time.Minute).Unix(),
	})

	accessTokenStr, err = accessToken.SignedString([]byte(Key))
	if err != nil {
		return "", "", err
	}

	//Генерирую рефреш токен
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(60 * time.Minute).Unix(),
	})

	refreshTokenStr, err = refreshToken.SignedString([]byte(Key))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, err
}

// Регенарция Токенов
func RegenerateTokens(refreshTokenStr string) (newAccessTokenStr, newRefreshTokenStr string, err error) {
	Key := Configs.EnvConfigs.JwtToken
	// Достаю логин из рефреш токена, чтобы для этого же логина перегенерировать токены
	token, err := jwt.ParseWithClaims(refreshTokenStr, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", err
	}

	login, ok := (*claims)["login"].(string)
	if !ok {
		return "", "", err
	}

	// Генерация нового аксесс токена
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(1 * time.Minute).Unix(),
	})
	newAccessTokenStr, err = newAccessToken.SignedString([]byte(Key))
	if err != nil {
		return "", "", err
	}

	// Генерация нового рефреш токена
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(60 * time.Minute).Unix(),
	})
	newRefreshTokenStr, err = newRefreshToken.SignedString([]byte(Key))
	if err != nil {
		return "", "", err
	}

	return newAccessTokenStr, newRefreshTokenStr, nil
}

// Присваивание куков
func SetTokenCookies(w http.ResponseWriter, login string) error {

	accessToken, refreshToken, err := GenerateTokens(login)
	if err != nil {
		fmt.Println("Ошибка генерации токенов")
		return err
	} else {

		accessCookie := http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Expires:  time.Now().Add(1 * time.Minute),
			HttpOnly: true,
		}

		refreshCookie := http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Expires:  time.Now().Add(60 * time.Minute),
			HttpOnly: true,
		}

		http.SetCookie(w, &accessCookie)
		http.SetCookie(w, &refreshCookie)
		fmt.Println("Успешно присвоены куки")
		w.Write([]byte("Акссес и рефреш токены присвоены успешно!"))
		return nil
	}
}

// Присваивание куков новых токенов
func SetNewTokensCookies(w http.ResponseWriter, accessTokenStr string, refreshTokenStr string) error {

	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessTokenStr,
		Expires:  time.Now().Add(1 * time.Minute),
		HttpOnly: true,
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenStr,
		Expires:  time.Now().Add(60 * time.Minute),
		HttpOnly: true,
	}

	http.SetCookie(w, &accessCookie)
	http.SetCookie(w, &refreshCookie)
	fmt.Println("Успешно присвоены новые куки")
	w.Write([]byte("Новые акссес и рефреш токены присвоены успешно!"))
	return nil

}

// Проверка на соответствие токена
func VerifyToken(tokenStr string) (bool, error) {
	Key := Configs.EnvConfigs.JwtToken
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Токен не соответствует необходимым параметрам: %v", token.Header["alg"])
		}
		// Здесь должен быть ваш секретный ключ
		return []byte(Key), nil
	})

	if err != nil {
		return false, err
	}

	return token.Valid, nil
}

// Проверка, ействует ли ещё токен
func IsTokenExpired(tokenStr string) (bool, error) {
	Key := Configs.EnvConfigs.JwtToken

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})
	if err != nil {
		return false, err // Невозможно распарсить токен
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			now := time.Now().Unix()
			return now > int64(exp), nil
		}
	}

	return false, fmt.Errorf("Токен не соответствует необходимым параметрам")
}

// Получение Email из структуры
func GetUserEmail(AccessTokenStr string) string {

	Configs.InitYamlConfig()
	var users []map[string]string
	if err := viper.UnmarshalKey("users", &users); err != nil {
		return ""
	}

	login, err := getLoginFromToken(AccessTokenStr)
	if err == nil {
		for _, user := range users {
			if user["login"] == login {

				return user["email"]
			} else {
				return "У данного пользователя нет Email"
			}
		}
	}
	return ""
}

// Получение Логина из токена
func getLoginFromToken(tokenString string) (string, error) {
	Key := Configs.EnvConfigs.JwtToken

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(Key), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if login, ok := claims["login"].(string); ok {
			return login, nil
		}
		return "", fmt.Errorf("Логин не найден в токене")
	} else {
		return "", fmt.Errorf("Невозможно обработать токен")
	}
}
