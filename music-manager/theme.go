package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type SpotifyTheme struct{}

func (SpotifyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{18, 18, 18, 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{40, 40, 40, 255}
	case theme.ColorNameButton, theme.ColorNamePrimary:
		return color.NRGBA{30, 215, 96, 255}
	case theme.ColorNameForeground:
		return color.NRGBA{230, 230, 230, 255}
	}
	return theme.DefaultTheme().Color(n, v)
}
func (SpotifyTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (SpotifyTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (SpotifyTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }
