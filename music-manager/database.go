package main

import (
	"database/sql"

	"fyne.io/fyne/v2"
)

var (
	db          *sql.DB // Соединение с базой данных
	repo        *Repository
	currentUser *User // Текущий авторизованный пользователь
	mainWindow  fyne.Window
)
