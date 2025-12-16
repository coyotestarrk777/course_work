package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	// Импорты Fyne
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	// Импорты для БД
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// --- СТРУКТУРЫ ДАННЫХ ---
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

type Track struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	AlbumID  int    `json:"album_id"`
	Duration int    `json:"duration"`
}

// --- ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ ---
var db *sql.DB
var currentUser *User // Для отслеживания текущего вошедшего пользователя
var mainWindow fyne.Window // Главное окно для диалогов

// --- 1. ФУНКЦИИ РАБОТЫ С БАЗОЙ ДАННЫХ (CRUD) ---

// Хелпер для получения всех треков
func getTracksFromDB() ([]Track, error) {
	rows, err := db.Query("SELECT id, title, album_id, duration FROM tracks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса треков: %w", err)
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var t Track
		if err := rows.Scan(&t.ID, &t.Title, &t.AlbumID, &t.Duration); err != nil {
			log.Printf("Error scanning track: %v", err)
			continue
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

// Хелпер для создания трека
func createTrackInDB(t Track) error {
	_, err := db.Exec("INSERT INTO tracks (title, album_id, duration) VALUES ($1, $2, $3)",
		t.Title, t.AlbumID, t.Duration)
	return err
}

// Хелпер для удаления трека
func deleteTrackInDB(id int) error {
	result, err := db.Exec("DELETE FROM tracks WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("трек с ID %d не найден", id)
	}
	return nil
}

// --- 2. ФУНКЦИИ АУТЕНТИФИКАЦИИ ---

func registerUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username, string(hashedPassword))

	if err != nil {
		return fmt.Errorf("регистрация не удалась: %w", err)
	}
	return nil
}

func loginUser(username, password string) error {
	var storedHash string
	var userID int
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = $1", username).Scan(&userID, &storedHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь не найден")
		}
		return fmt.Errorf("ошибка базы данных при входе: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return fmt.Errorf("неверный пароль")
	}

	currentUser = &User{ID: userID, Username: username}
	return nil
}

// --- 3. ФУНКЦИИ UI (Fyne) ---

// Функция для создания вкладки "Управление треками" (CRUD)
func createTrackCRUDTab(updateTabs func()) *container.TabItem {
	// 1. Поле вывода списка треков
	tracksList := widget.NewLabel("Нажмите 'Обновить', чтобы увидеть треки")
	tracksList.Wrapping = fyne.TextWrapBreak // Перенос текста

	// Функция для обновления списка треков
	updateTracksList := func() {
		tracks, err := getTracksFromDB()
		if err != nil {
			tracksList.SetText(fmt.Sprintf("Ошибка загрузки: %v", err))
			return
		}

		if len(tracks) == 0 {
			tracksList.SetText("В базе данных нет треков.")
			return
		}

		output := "ID | Название | Альбом ID | Длительность (с)\n"
		output += "--------------------------------------------------------\n"
		for _, t := range tracks {
			output += fmt.Sprintf("%d | %s | %d | %d\n", t.ID, t.Title, t.AlbumID, t.Duration)
		}
		tracksList.SetText(output)
	}

	// 2. Форма добавления трека
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название трека")
	albumIDEntry := widget.NewEntry()
	albumIDEntry.SetPlaceHolder("ID Альбома (число)")
	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("Длительность (секунды)")

	createButton := widget.NewButton("Создать трек", func() {
		albumID, err1 := strconv.Atoi(albumIDEntry.Text)
		duration, err2 := strconv.Atoi(durationEntry.Text)

		if titleEntry.Text == "" || err1 != nil || err2 != nil {
			dialog.ShowError(fmt.Errorf("пожалуйста, заполните все поля корректно"), mainWindow)
			return
		}

		track := Track{Title: titleEntry.Text, AlbumID: albumID, Duration: duration}
		err := createTrackInDB(track)
		if err != nil {
			dialog.ShowError(fmt.Errorf("не удалось создать трек: %w", err), mainWindow)
		} else {
			dialog.ShowInformation("Успех", "Трек успешно создан!", mainWindow)
			updateTracksList()
			titleEntry.SetText("")
			albumIDEntry.SetText("")
			durationEntry.SetText("")
		}
	})

	// 3. Форма удаления трека
	deleteIDEntry := widget.NewEntry()
	deleteIDEntry.SetPlaceHolder("ID трека для удаления")
	deleteButton := widget.NewButton("Удалить трек", func() {
		id, err := strconv.Atoi(deleteIDEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("введите корректный ID"), mainWindow)
			return
		}

		err = deleteTrackInDB(id)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка удаления: %w", err), mainWindow)
		} else {
			dialog.ShowInformation("Успех", "Трек успешно удален!", mainWindow)
			updateTracksList()
			deleteIDEntry.SetText("")
		}
	})

	// Сборка вкладки
	content := container.NewVBox(
		widget.NewButton("Обновить список треков", updateTracksList),
		widget.NewSeparator(),
		container.NewVScroll(tracksList),

		widget.NewSeparator(),
		widget.NewLabel("ДОБАВИТЬ НОВЫЙ ТРЕК:"),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), titleEntry, albumIDEntry, durationEntry),
		createButton,

		widget.NewSeparator(),
		widget.NewLabel("УДАЛИТЬ ТРЕК:"),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(150, 30)), deleteIDEntry, deleteButton),
	)

	return container.NewTabItem("🎵 Треки (CRUD)", content)
}

// Функция для создания контейнера аутентификации
func createAuthUI(a fyne.App, showContent func()) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Имя пользователя")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Подтвердите пароль (для регистрации)")
	confirmPasswordEntry.Hide() // Скрываем по умолчанию

	statusLabel := widget.NewLabel("Введите данные для входа или регистрации")

	// Кнопки для переключения
	loginMode := true
	registerLink := widget.NewHyperlink("Нет аккаунта? Зарегистрироваться", nil)
	loginLink := widget.NewHyperlink("Уже есть аккаунт? Войти", nil)
	loginLink.Hide()

	authButton := widget.NewButton("Войти", nil) // Заглушка, функция будет установлена ниже

	// Функция переключения режима
	toggleMode := func(toRegister bool) {
		loginMode = !toRegister
		if loginMode {
			authButton.SetText("Войти")
			registerLink.Show()
			loginLink.Hide()
			confirmPasswordEntry.Hide()
		} else {
			authButton.SetText("Зарегистрироваться")
			registerLink.Hide()
			loginLink.Show()
			confirmPasswordEntry.Show()
		}
	}

	// Обработчик кнопки Войти/Зарегистрироваться
	authButton.OnTapped = func() {
		statusLabel.SetText("Обработка...")

		if loginMode {
			// Логика входа
			err := loginUser(usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("❌ Ошибка входа: %v", err))
			} else {
				statusLabel.SetText(fmt.Sprintf("✅ Успешный вход! Добро пожаловать, %s!", currentUser.Username))
				showContent() // Показываем основное содержимое
			}
		} else {
			// Логика регистрации
			if passwordEntry.Text != confirmPasswordEntry.Text {
				statusLabel.SetText("❌ Пароли не совпадают!")
				return
			}
			err := registerUser(usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("❌ Ошибка регистрации: %v", err))
			} else {
				// После успешной регистрации пытаемся войти
				loginErr := loginUser(usernameEntry.Text, passwordEntry.Text)
				if loginErr != nil {
					statusLabel.SetText("✅ Регистрация успешна, но вход не удался. Попробуйте войти.")
					toggleMode(false) // Возвращаемся в режим входа
				} else {
					statusLabel.SetText(fmt.Sprintf("✅ Регистрация и вход успешны! Добро пожаловать, %s!", currentUser.Username))
					showContent() // Показываем основное содержимое
				}
			}
		}
	}

	registerLink.OnTapped = func() { toggleMode(true) }
	loginLink.OnTapped = func() { toggleMode(false) }

	return container.NewVBox(
		widget.NewLabelWithStyle("Вход / Регистрация", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		usernameEntry,
		passwordEntry,
		confirmPasswordEntry,
		authButton,
		container.NewHBox(registerLink, loginLink),
		widget.NewSeparator(),
		statusLabel,
	)
}

// Функция для создания основного содержимого (после входа)
func createMainContent(a fyne.App) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		createTrackCRUDTab(func() {}), // Вкладка CRUD
		// Здесь вы можете добавить createAlbumCRUDTab, createArtistCRUDTab и т.д.
	)

	// Добавляем заголовок с именем пользователя и кнопкой выхода
	header := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Управление треками"),
		widget.NewButtonWithIcon("Выход", theme.LogoutIcon(), func() {
			currentUser = nil
			mainWindow.SetContent(createAuthUI(a, func() {
				mainWindow.SetContent(createMainContent(a))
			}))
		}),
		tabs,
	)

	return header
}

// --- 4. MAIN FUNCTION ---

func main() {
	// 1. Подключение к базе данных
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database. Is Docker/PostgreSQL running? Details:", err)
	}
	fmt.Println("Успешное подключение к базе данных!")

	// 2. Инициализация и запуск Fyne приложения
	a := app.New()
	mainWindow = a.NewWindow("Музыкальный Каталог (Desktop)")
	mainWindow.Resize(fyne.NewSize(800, 600))
	mainWindow.SetMaster() // Закрытие главного окна завершает приложение

	// Функция для переключения на основное содержимое
	showContent := func() {
		mainWindow.SetContent(createMainContent(a))
	}

	// Устанавливаем начальный контент: форму аутентификации
	authContent := createAuthUI(a, showContent)
	
	// Чтобы центрировать форму аутентификации
	centeredAuth := container.NewCenter(authContent)

	mainWindow.SetContent(centeredAuth)
	
	// Устанавливаем логотип (опционально)
	mainWindow.SetIcon(theme.FolderIcon()) // Замените на свой значок

	mainWindow.ShowAndRun()
}