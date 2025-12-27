package main

import (
	"fmt"
)

// Регистрация пользователя
func registerUser(u, p string) error {
	// Базовая проверка входных данных остается в логике
	if u == "" || p == "" {
		return fmt.Errorf("логин и пароль не могут быть пустыми")
	}

	// Вызываем метод репозитория.
	// Репозиторий сам захеширует пароль и выполнит INSERT.
	return repo.RegisterUser(u, p)
}

// Вход пользователя
func loginUser(u, p string) error {
	// Вызываем метод репозитория.
	// Он проверит существование пользователя и совпадение хеша пароля.
	user, err := repo.LoginUser(u, p)
	if err != nil {
		return err
	}

	// Если ошибок нет, сохраняем пользователя в глобальную переменную
	currentUser = user
	return nil
}
