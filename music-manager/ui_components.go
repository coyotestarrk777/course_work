package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PLAYLIST TAB
func createPlaylistTab() *container.TabItem {
	var playlists []Playlist
	var playlistNames []string
	var allTracksCached []Track
	var filteredTracks []Track
	var filteredTrackNames []string
	var playlistTracks []Track

	var selectedPlaylist *Playlist
	var selectedTrack *Track

	var list *widget.List
	var trackSelect *widget.Select
	var playlistSelect *widget.Select

	searchTrack := widget.NewEntry()
	searchTrack.SetPlaceHolder("Поиск трека для добавления...")

	// ФУНКЦИЯ ОБНОВЛЕНИЯ (Refresh)
	refresh := func() {
		playlists, playlistNames = getPlaylists()
		if len(allTracksCached) == 0 {
			allTracksCached, _ = getTracks()
		}

		filteredTracks = nil
		filteredTrackNames = nil
		searchText := strings.ToLower(searchTrack.Text)
		for _, t := range allTracksCached {
			if searchText == "" || strings.Contains(strings.ToLower(t.Title), searchText) {
				filteredTracks = append(filteredTracks, t)
				filteredTrackNames = append(filteredTrackNames, t.Title)
			}
		}

		playlistSelect.Options = playlistNames
		trackSelect.Options = filteredTrackNames

		if selectedPlaylist != nil {
			// Принимаем два значения: список объектов и список имен
			playlistTracks, _ = getTracksFromPlaylist(selectedPlaylist.ID)
		} else {
			playlistTracks = nil
		}

		playlistSelect.Refresh()
		trackSelect.Refresh()
		list.Refresh()
	}

	playlistSelect = widget.NewSelect(nil, func(s string) {
		for _, p := range playlists {
			if p.Title == s {
				selectedPlaylist = &p
				playlistTracks, _ = getTracksFromPlaylist(p.ID)
				break
			}
		}
		list.Refresh()
	})
	playlistSelect.PlaceHolder = "Выберите плейлист"

	// КНОПКА УДАЛЕНИЯ ПЛЕЙЛИСТА (Теперь она здесь)
	deletePlaylistBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if selectedPlaylist == nil {
			dialog.ShowInformation("Внимание", "Выберите плейлист для удаления", mainWindow)
			return
		}
		confirmDelete("Удаление", "Удалить плейлист '"+selectedPlaylist.Title+"'?", func() {
			deletePlaylist(selectedPlaylist.ID)
			selectedPlaylist = nil
			playlistSelect.ClearSelected()
			refresh()
		})
	})
	deletePlaylistBtn.Importance = widget.DangerImportance

	trackSelect = widget.NewSelect(nil, func(s string) {
		for _, t := range filteredTracks {
			if t.Title == s {
				selectedTrack = &t
				break
			}
		}
	})
	trackSelect.PlaceHolder = "Выберите трек"

	addTrackBtn := widget.NewButtonWithIcon("Добавить в плейлист", theme.ContentAddIcon(), func() {
		if selectedPlaylist == nil || selectedTrack == nil {
			dialog.ShowInformation("Внимание", "Выберите плейлист и трек", mainWindow)
			return
		}
		err := repo.AddTrackToPlaylist(selectedPlaylist.ID, selectedTrack.ID)
		if err != nil {
			dialog.ShowError(err, mainWindow)
		}
		refresh()
	})

	newPlaylistEntry := widget.NewEntry()
	newPlaylistEntry.SetPlaceHolder("Название нового плейлиста")
	addPlaylistBtn := widget.NewButtonWithIcon("Создать плейлист", theme.DocumentCreateIcon(), func() {
		if newPlaylistEntry.Text != "" {
			createPlaylist(newPlaylistEntry.Text)
			newPlaylistEntry.SetText("")
			refresh()
		}
	})

	list = widget.NewList(
		func() int { return len(playlistTracks) },
		func() fyne.CanvasObject {
			return listRowWithDelete("Название трека", func() {})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(playlistTracks) {
				return
			}
			track := playlistTracks[i]
			min := track.Duration / 60
			sec := track.Duration % 60
			title := fmt.Sprintf("%s (%d:%02d)", track.Title, min, sec)
			o.(*fyne.Container).Objects[0].(*widget.Label).SetText(title)
			o.(*fyne.Container).Objects[2].(*widget.Button).OnTapped = func() {
				confirmDelete("Удаление", "Удалить трек из плейлиста?", func() {
					repo.RemoveTrackFromPlaylist(selectedPlaylist.ID, track.ID)
					refresh()
				})
			}
		},
	)

	searchTrack.OnChanged = func(string) { refresh() }

	// Компоновка верхней части (Селектор + Кнопка удаления в одной строке)
	playlistHeader := container.NewBorder(nil, nil, nil, deletePlaylistBtn, playlistSelect)

	refresh()

	return container.NewTabItemWithIcon("Плейлисты", theme.StorageIcon(), container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Управление плейлистами", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewBorder(nil, nil, nil, addPlaylistBtn, newPlaylistEntry),
			widget.NewSeparator(),
			widget.NewLabel("Текущий плейлист:"),
			playlistHeader,
			widget.NewSeparator(),
			widget.NewLabel("Добавить треки:"),
			searchTrack,
			container.NewBorder(nil, nil, nil, addTrackBtn, trackSelect),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		list,
	))
}

// DATABASE TAB
func createDatabaseTab() *container.TabItem {
	var artists []Artist
	var artistNames []string
	var albums []Album
	var albumNames []string
	var tracks []Track
	var trackNames []string

	artistList := widget.NewList(nil, nil, nil)
	albumList := widget.NewList(nil, nil, nil)
	trackList := widget.NewList(nil, nil, nil)

	newArtistEntry := widget.NewEntry()
	newArtistEntry.SetPlaceHolder("Имя артиста")
	newAlbumEntry := widget.NewEntry()
	newAlbumEntry.SetPlaceHolder("Название альбома")
	newAlbumYearEntry := widget.NewEntry()
	newAlbumYearEntry.SetPlaceHolder("Год (напр. 2023)")
	newTrackEntry := widget.NewEntry()
	newTrackEntry.SetPlaceHolder("Название трека")
	newTrackDurationEntry := widget.NewEntry()
	newTrackDurationEntry.SetPlaceHolder("Секунды")

	searchArtist := widget.NewEntry()
	searchArtist.SetPlaceHolder("Поиск...")
	searchAlbum := widget.NewEntry()
	searchAlbum.SetPlaceHolder("Поиск...")
	searchTrack := widget.NewEntry()
	searchTrack.SetPlaceHolder("Поиск...")

	albumSelectArtist := widget.NewSelect(nil, nil)
	trackSelectAlbum := widget.NewSelect(nil, nil)

	refreshAll := func() {
		// Артисты
		allA, allAN := getArtists()
		artists, artistNames = nil, nil
		for i, n := range allAN {
			if containsIgnoreCase(n, searchArtist.Text) {
				artists = append(artists, allA[i])
				artistNames = append(artistNames, n)
			}
		}
		albumSelectArtist.Options = allAN

		// Альбомы
		allAlb, allAlbN := getAlbums()
		albums, albumNames = nil, nil
		for i, n := range allAlbN {
			if containsIgnoreCase(n, searchAlbum.Text) {
				albums = append(albums, allAlb[i])
				albumNames = append(albumNames, n)
			}
		}
		trackSelectAlbum.Options = allAlbN

		// Треки
		allT, allTN := getTracks()
		tracks, trackNames = nil, nil
		for i, n := range allTN {
			if containsIgnoreCase(n, searchTrack.Text) {
				tracks = append(tracks, allT[i])
				trackNames = append(trackNames, n)
			}
		}

		artistList.Refresh()
		albumList.Refresh()
		trackList.Refresh()
	}

	// Настройка списков
	artistList.Length = func() int { return len(artistNames) }
	artistList.CreateItem = func() fyne.CanvasObject { return listRowWithDelete("", func() {}) }
	artistList.UpdateItem = func(id widget.ListItemID, o fyne.CanvasObject) {
		if id >= len(artists) {
			return
		}
		a := artists[id]
		o.(*fyne.Container).Objects[0].(*widget.Label).SetText(a.Name)
		o.(*fyne.Container).Objects[2].(*widget.Button).OnTapped = func() {
			confirmDelete("Удаление", "Удалить артиста "+a.Name+"?", func() {
				deleteArtist(a.ID)
				refreshAll()
			})
		}
	}
	albumList.Length = func() int { return len(albumNames) }
	albumList.CreateItem = func() fyne.CanvasObject { return listRowWithDelete("", func() {}) }
	albumList.UpdateItem = func(id widget.ListItemID, o fyne.CanvasObject) {
		if id >= len(albums) {
			return
		}
		a := albums[id]
		o.(*fyne.Container).Objects[0].(*widget.Label).SetText(albumNames[id])
		o.(*fyne.Container).Objects[2].(*widget.Button).OnTapped = func() {
			confirmDelete("Удаление", "Удалить альбом?", func() {
				deleteAlbum(a.ID)
				refreshAll()
			})
		}
	}

	trackList.Length = func() int { return len(trackNames) }
	trackList.CreateItem = func() fyne.CanvasObject { return listRowWithDelete("", func() {}) }
	trackList.UpdateItem = func(id widget.ListItemID, o fyne.CanvasObject) {
		if id >= len(tracks) {
			return
		}
		t := tracks[id]
		o.(*fyne.Container).Objects[0].(*widget.Label).SetText(trackNames[id])
		o.(*fyne.Container).Objects[2].(*widget.Button).OnTapped = func() {
			confirmDelete("Удаление", "Удалить трек?", func() {
				deleteTrack(t.ID)
				refreshAll()
			})
		}
	}

	// Кнопки добавления
	addArtBtn := widget.NewButton("Добавить", func() {
		if newArtistEntry.Text != "" {
			addArtist(newArtistEntry.Text)
			newArtistEntry.SetText("")
			refreshAll()
		}
	})

	addAlbBtn := widget.NewButton("Добавить", func() {
		var artID int
		allA, _ := getArtists()
		for _, a := range allA {
			if a.Name == albumSelectArtist.Selected {
				artID = a.ID
				break
			}
		}
		year, _ := strconv.Atoi(newAlbumYearEntry.Text)
		addAlbum(newAlbumEntry.Text, artID, year)
		newAlbumEntry.SetText("")
		refreshAll()
	})

	addTrackBtn := widget.NewButton("Добавить", func() {
		var alID int
		allAl, _ := getAlbums()
		for _, a := range allAl {
			if fmt.Sprintf("%s (%d)", a.Title, a.Year) == trackSelectAlbum.Selected {
				alID = a.ID
				break
			}
		}
		dur, _ := strconv.Atoi(newTrackDurationEntry.Text)
		addTrack(newTrackEntry.Text, alID, dur)
		newTrackEntry.SetText("")
		refreshAll()
	})

	searchArtist.OnChanged = func(string) { refreshAll() }
	searchAlbum.OnChanged = func(string) { refreshAll() }
	searchTrack.OnChanged = func(string) { refreshAll() }

	refreshAll()

	return container.NewTabItemWithIcon("База данных", theme.InfoIcon(), container.NewAppTabs(
		container.NewTabItem("Артисты", container.NewBorder(container.NewVBox(newArtistEntry, addArtBtn, searchArtist), nil, nil, nil, artistList)),
		container.NewTabItem("Альбомы", container.NewBorder(container.NewVBox(albumSelectArtist, newAlbumEntry, newAlbumYearEntry, addAlbBtn, searchAlbum), nil, nil, nil, albumList)),
		container.NewTabItem("Треки", container.NewBorder(container.NewVBox(trackSelectAlbum, newTrackEntry, newTrackDurationEntry, addTrackBtn, searchTrack), nil, nil, nil, trackList)),
	))
}
