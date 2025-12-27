package main

import (
	"image/color"

	"fyne.io/fyne/v2"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	_ "github.com/lib/pq"
)

func createAuthUI(onSuccess func()) fyne.CanvasObject {
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Логин")
	userEntry.TextStyle = fyne.TextStyle{Bold: true}
	userEntry.Resize(fyne.NewSize(300, 40))

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Пароль")
	passEntry.TextStyle = fyne.TextStyle{Bold: true}
	passEntry.Resize(fyne.NewSize(300, 40))

	title := canvas.NewText("MUSIC MANAGER", color.NRGBA{30, 215, 96, 255})
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}

	loginBtn := widget.NewButton("Войти", func() {
		err := loginUser(userEntry.Text, passEntry.Text)
		if err != nil {
			dialog.ShowError(err, mainWindow)
		} else {
			onSuccess()
		}
	})

	regBtn := widget.NewButton("Регистрация", func() {
		err := registerUser(userEntry.Text, passEntry.Text)
		if err != nil {
			dialog.ShowError(err, mainWindow)
		} else {
			dialog.ShowInformation("Успех", "Аккаунт создан. Теперь можно войти.", mainWindow)
		}
	})

	form := container.NewVBox(
		container.NewCenter(title),
		userEntry,
		passEntry,
		loginBtn,
		regBtn,
	)

	return container.NewCenter(form)
}
