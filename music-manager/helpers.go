package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	_ "github.com/lib/pq"
)

func confirmDelete(title, message string, onDelete func()) { // окно для подтверждения удаления
	dialog.ShowConfirm(title, message, func(ok bool) {
		if ok {
			onDelete()
		}
	}, mainWindow)
}

// создание кнопки удаления
func listRowWithDelete(title string, onDelete func()) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		confirmDelete("Удаление", fmt.Sprintf("Вы уверены, что хотите удалить %s?", title), onDelete)
	})
	deleteBtn.Importance = widget.LowImportance
	return container.NewHBox(label, layout.NewSpacer(), deleteBtn)
}

func execInsert(q string, args ...interface{}) error { // функция для insertзапросов
	_, err := db.Exec(q, args...)
	return err
}

// проверка строки на подстроку
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
