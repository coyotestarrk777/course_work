package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Формируем строку подключения из переменных окружения
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	var err error
	// 2. Открываем соединение с БД
	db, err = sql.Open("postgres", conn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// 3. ИНИЦИАЛИЗИРУЕМ РЕПОЗИТОРИЙ (Важное изменение)
	// Теперь переменная 'repo' из database.go заполнена и готова к работе
	repo = NewRepository(db)

	// 4. Создаем приложение и настраиваем тему
	myApp := app.New()
	myApp.Settings().SetTheme(&SpotifyTheme{})

	// 5. Создаем главное окно
	mainWindow = myApp.NewWindow("Music Manager")

	// Устанавливаем стартовый экран (Авторизация)
	mainWindow.SetContent(createAuthUI(func() {
		// При успешном входе переключаемся на основной интерфейс
		mainWindow.SetContent(container.NewAppTabs(
			createPlaylistTab(),
			createDatabaseTab(),
		))
	}))

	mainWindow.ShowAndRun()
}
