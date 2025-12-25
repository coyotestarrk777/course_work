package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ================= THEME =================
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

// ================= DATA STRUCTURES =================
type User struct {
	ID       int
	Username string
}

type Playlist struct {
	ID    int
	Title string
}

type Artist struct {
	ID   int
	Name string
}

type Album struct {
	ID       int
	Title    string
	ArtistID int
}

type Track struct {
	ID      int
	Title   string
	AlbumID int
}

var (
	db          *sql.DB
	currentUser *User
	mainWindow  fyne.Window
)

// ================= HELPERS =================
func confirmDelete(title, message string, onDelete func()) {
	dialog.ShowConfirm(title, message, func(ok bool) {
		if ok {
			onDelete()
		}
	}, mainWindow)
}

func listRowWithDelete(title string, onDelete func()) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		confirmDelete("Удаление", fmt.Sprintf("Вы уверены, что хотите удалить %s?", title), onDelete)
	})
	deleteBtn.Importance = widget.LowImportance
	return container.NewHBox(label, layout.NewSpacer(), deleteBtn)
}

func execInsert(q string, args ...interface{}) error {
	_, err := db.Exec(q, args...)
	return err
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ================= DB OPERATIONS =================
func registerUser(u, p string) error {
	if u == "" || p == "" {
		return fmt.Errorf("логин и пароль не могут быть пустыми")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return execInsert("INSERT INTO users (username,password_hash) VALUES ($1,$2)", u, string(hash))
}

func loginUser(u, p string) error {
	var id int
	var hash string
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username=$1", u).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) != nil {
		return fmt.Errorf("неверный логин или пароль")
	}
	currentUser = &User{ID: id, Username: u}
	return nil
}

// --- PLAYLISTS ---
func getPlaylists() ([]Playlist, []string) {
	rows, err := db.Query("SELECT id, title FROM playlists WHERE user_id=$1 ORDER BY title", currentUser.ID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	var items []Playlist
	var names []string
	for rows.Next() {
		var p Playlist
		rows.Scan(&p.ID, &p.Title)
		items = append(items, p)
		names = append(names, p.Title)
	}
	return items, names
}

func createPlaylist(title string) error {
	if title == "" {
		return fmt.Errorf("название плейлиста не может быть пустым")
	}
	return execInsert("INSERT INTO playlists (title, user_id) VALUES ($1,$2)", title, currentUser.ID)
}

func deletePlaylist(playlistID int) error {
	_, err := db.Exec(`DELETE FROM playlist_tracks WHERE playlist_id=$1; DELETE FROM playlists WHERE id=$1`, playlistID)
	return err
}

// --- ARTISTS ---
func addArtist(name string) error {
	if name == "" {
		return fmt.Errorf("название артиста не может быть пустым")
	}
	return execInsert("INSERT INTO artists (name) VALUES ($1)", name)
}

func getArtists() ([]Artist, []string) {
	rows, _ := db.Query("SELECT id, name FROM artists ORDER BY name")
	defer rows.Close()
	var items []Artist
	var names []string
	for rows.Next() {
		var a Artist
		rows.Scan(&a.ID, &a.Name)
		items = append(items, a)
		names = append(names, a.Name)
	}
	return items, names
}

func deleteArtist(artistID int) error {
	_, err := db.Exec(`
    DELETE FROM playlist_tracks
    WHERE track_id IN (
      SELECT t.id FROM tracks t
      JOIN albums a ON t.album_id = a.id
      WHERE a.artist_id = $1
    );

    DELETE FROM tracks
    WHERE album_id IN (SELECT id FROM albums WHERE artist_id = $1);

    DELETE FROM albums WHERE artist_id = $1;
    DELETE FROM artists WHERE id = $1;
  `, artistID)
	return err
}

// --- ALBUMS ---
func addAlbum(title string, artistID int) error {
	if title == "" || artistID == 0 {
		return fmt.Errorf("альбом и артист должны быть выбраны")
	}
	return execInsert("INSERT INTO albums (title, artist_id) VALUES ($1,$2)", title, artistID)
}

func getAlbums() ([]Album, []string) {
	rows, _ := db.Query("SELECT id, title, artist_id FROM albums ORDER BY title")
	defer rows.Close()
	var items []Album
	var names []string
	for rows.Next() {
		var a Album
		rows.Scan(&a.ID, &a.Title, &a.ArtistID)
		items = append(items, a)
		names = append(names, a.Title)
	}
	return items, names
}

func deleteAlbum(albumID int) error {
	_, err := db.Exec(`
    DELETE FROM playlist_tracks
    WHERE track_id IN (SELECT id FROM tracks WHERE album_id = $1);

    DELETE FROM tracks WHERE album_id = $1;
    DELETE FROM albums WHERE id = $1;
  `, albumID)
	return err
}

// --- TRACKS ---
func addTrack(title string, albumID int) error {
	if title == "" || albumID == 0 {
		return fmt.Errorf("трек и альбом должны быть выбраны")
	}
	return execInsert("INSERT INTO tracks (title, album_id) VALUES ($1,$2)", title, albumID)
}

func getTracks() ([]Track, []string) {
	rows, _ := db.Query("SELECT id, title, album_id FROM tracks ORDER BY title")
	defer rows.Close()
	var items []Track
	var names []string
	for rows.Next() {
		var t Track
		rows.Scan(&t.ID, &t.Title, &t.AlbumID)
		items = append(items, t)
		names = append(names, t.Title)
	}
	return items, names
}

func deleteTrack(trackID int) error {
	_, err := db.Exec(`
    DELETE FROM playlist_tracks WHERE track_id = $1;
    DELETE FROM tracks WHERE id = $1;
  `, trackID)
	return err
}

// --- PLAYLIST TRACKS ---
func getTracksFromPlaylist(playlistID int) []Track {
	rows, _ := db.Query(`
		SELECT t.id, t.title
		FROM tracks t
		JOIN playlist_tracks pt ON pt.track_id = t.id
		WHERE pt.playlist_id = $1
		ORDER BY t.title
	`, playlistID)
	defer rows.Close()
	var tracks []Track
	for rows.Next() {
		var t Track
		rows.Scan(&t.ID, &t.Title)
		tracks = append(tracks, t)
	}
	return tracks
}

// ================= UI =================
// --- PLAYLIST TAB ---
func createPlaylistTab() *container.TabItem {
	playlists, playlistNames := getPlaylists()
	tracks, trackNames := getTracks()
	var selectedPlaylist *Playlist
	var selectedTrack *Track
	var playlistTracks []Track

	var list *widget.List

	loadTracks := func() {
		if selectedPlaylist != nil {
			playlistTracks = getTracksFromPlaylist(selectedPlaylist.ID)
			if list != nil {
				list.Refresh()
			}
		}
	}

	list = widget.NewList(
		func() int { return len(playlistTracks) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Название трека"),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(playlistTracks[i].Title)
			row.Objects[2].(*widget.Button).OnTapped = func() {
				if selectedPlaylist != nil {
					confirmDelete("Удаление трека из плейлиста", "Вы уверены, что хотите удалить трек из плейлиста?", func() {
						db.Exec("DELETE FROM playlist_tracks WHERE playlist_id=$1 AND track_id=$2",
							selectedPlaylist.ID, playlistTracks[i].ID)
						loadTracks()
					})
				}
			}
		},
	)

	listScroll := container.NewVScroll(list)
	listScroll.SetMinSize(fyne.NewSize(600, 400))

	playlistSelect := widget.NewSelect(playlistNames, func(s string) {
		for i, p := range playlists {
			if p.Title == s {
				selectedPlaylist = &playlists[i]
				break
			}
		}
		loadTracks()
	})

	trackSelect := widget.NewSelect(trackNames, func(s string) {
		for i, t := range tracks {
			if t.Title == s {
				selectedTrack = &tracks[i]
				break
			}
		}
	})

	addBtn := widget.NewButtonWithIcon("Добавить трек", theme.ContentAddIcon(), func() {
		if selectedPlaylist == nil || selectedTrack == nil {
			dialog.ShowInformation("Внимание", "Выберите плейлист и трек", mainWindow)
			return
		}
		err := execInsert("INSERT INTO playlist_tracks (playlist_id, track_id) VALUES ($1, $2)",
			selectedPlaylist.ID, selectedTrack.ID)
		if err != nil {
			dialog.ShowError(err, mainWindow)
		} else {
			loadTracks()
		}
	})

	newPlaylistEntry := widget.NewEntry()
	newPlaylistEntry.SetPlaceHolder("Название нового плейлиста")
	newPlaylistEntry.TextStyle = fyne.TextStyle{Bold: true}
	newPlaylistEntry.Resize(fyne.NewSize(300, 40))

	newPlaylistBtn := widget.NewButton("Создать плейлист", func() {
		if err := createPlaylist(newPlaylistEntry.Text); err != nil {
			dialog.ShowError(err, mainWindow)
		} else {
			playlists, playlistNames = getPlaylists()
			playlistSelect.Options = playlistNames
			playlistSelect.Refresh()
			newPlaylistEntry.SetText("")
		}
	})

	searchPlaylistTrackEntry := widget.NewEntry()
	searchPlaylistTrackEntry.SetPlaceHolder("Поиск трека в плейлисте...")
	searchPlaylistTrackEntry.OnChanged = func(s string) {
		if selectedPlaylist != nil {
			allTracks := getTracksFromPlaylist(selectedPlaylist.ID)
			var filtered []Track
			for _, t := range allTracks {
				if s == "" || containsIgnoreCase(t.Title, s) {
					filtered = append(filtered, t)
				}
			}
			playlistTracks = filtered
			list.Refresh()
		}
	}

	content := container.NewBorder(
		container.NewVBox(
			newPlaylistEntry,
			newPlaylistBtn,
			playlistSelect,
			trackSelect,
			addBtn,
			searchPlaylistTrackEntry,
		),
		nil, nil, nil,
		listScroll,
	)

	return container.NewTabItemWithIcon("Плейлисты", theme.FolderIcon(), content)
}

// --- DATABASE TAB ---
func createDatabaseTab() *container.TabItem {
	artists, artistNames := getArtists()
	albums, albumNames := getAlbums()
	tracks, trackNames := getTracks()

	// --- ARTISTS ---
	var artistList *widget.List
	searchArtistEntry := widget.NewEntry()
	searchArtistEntry.SetPlaceHolder("Поиск артиста...")
	searchArtistEntry.OnChanged = func(s string) {
		var filteredNames []string
		for _, a := range artists {
			if s == "" || containsIgnoreCase(a.Name, s) {
				filteredNames = append(filteredNames, a.Name)
			}
		}
		artistNames = filteredNames
		artistList.Refresh()
	}

	artistList = widget.NewList(
		func() int { return len(artistNames) },
		func() fyne.CanvasObject { return listRowWithDelete("Артист", func() {}) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(artistNames[i])
			row.Objects[2].(*widget.Button).OnTapped = func() {
				confirmDelete("Удаление артиста", "Будут удалены все альбомы и треки этого артиста. Продолжить?", func() {
					if err := deleteArtist(artists[i].ID); err != nil {
						dialog.ShowError(err, mainWindow)
						return
					}
					artists, artistNames = getArtists()
					albums, albumNames = getAlbums()
					tracks, trackNames = getTracks()
					artistList.Refresh()
				})
			}
		})

	artistScroll := container.NewVScroll(artistList)
	artistScroll.SetMinSize(fyne.NewSize(600, 200))

	newArtistEntry := widget.NewEntry()
	newArtistEntry.SetPlaceHolder("Название артиста")
	newArtistEntry.TextStyle = fyne.TextStyle{Bold: true}
	newArtistBtn := widget.NewButtonWithIcon("Добавить артиста", theme.ContentAddIcon(), func() {
		if err := addArtist(newArtistEntry.Text); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		artists, artistNames = getArtists()
		artistList.Refresh()
		newArtistEntry.SetText("")
	})

	artistTab := container.NewBorder(
		container.NewVBox(newArtistEntry, newArtistBtn, searchArtistEntry, widget.NewSeparator()),
		nil, nil, nil,
		artistScroll,
	)

	// --- ALBUMS ---
	var albumList *widget.List
	albumSelectArtist := widget.NewSelect(artistNames, nil)
	albumSelectArtist.PlaceHolder = "Выберите артиста"

	searchAlbumEntry := widget.NewEntry()
	searchAlbumEntry.SetPlaceHolder("Поиск альбома...")
	searchAlbumEntry.OnChanged = func(s string) {
		var filteredAlbums []Album
		var filteredAlbumNames []string
		for _, al := range albums {
			if s == "" || containsIgnoreCase(al.Title, s) {
				filteredAlbums = append(filteredAlbums, al)
				filteredAlbumNames = append(filteredAlbumNames, al.Title)
			}
		}
		albums = filteredAlbums
		albumNames = filteredAlbumNames
		albumList.Refresh()
	}

	albumList = widget.NewList(
		func() int { return len(albumNames) },
		func() fyne.CanvasObject { return listRowWithDelete("Альбом", func() {}) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(albumNames[i])
			row.Objects[2].(*widget.Button).OnTapped = func() {
				confirmDelete("Удаление альбома", "Будут удалены все треки этого альбома. Продолжить?", func() {
					if err := deleteAlbum(albums[i].ID); err != nil {
						dialog.ShowError(err, mainWindow)
						return
					}
					albums, albumNames = getAlbums()
					albumList.Refresh()
				})
			}
		})

	albumScroll := container.NewVScroll(albumList)
	albumScroll.SetMinSize(fyne.NewSize(600, 200))

	newAlbumEntry := widget.NewEntry()
	newAlbumEntry.SetPlaceHolder("Название альбома")
	newAlbumBtn := widget.NewButton("Добавить альбом", func() {
		if albumSelectArtist.Selected == "" || newAlbumEntry.Text == "" {
			dialog.ShowInformation("Внимание", "Выберите артиста и введите название альбома", mainWindow)
			return
		}
		var artistID int
		for _, a := range artists {
			if a.Name == albumSelectArtist.Selected {
				artistID = a.ID
				break
			}
		}
		if err := addAlbum(newAlbumEntry.Text, artistID); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		albums, albumNames = getAlbums()
		newAlbumEntry.SetText("")
		if albumSelectArtist.Selected != "" {
			albumSelectArtist.OnChanged(albumSelectArtist.Selected)
		}
	})

	albumSelectArtist.OnChanged = func(s string) {
		var selectedArtistID int
		for _, a := range artists {
			if a.Name == s {
				selectedArtistID = a.ID
				break
			}
		}
		var filteredAlbums []Album
		var filteredAlbumNames []string
		for _, al := range albums {
			if al.ArtistID == selectedArtistID {
				filteredAlbums = append(filteredAlbums, al)
				filteredAlbumNames = append(filteredAlbumNames, al.Title)
			}
		}
		albums = filteredAlbums
		albumNames = filteredAlbumNames
		albumList.Refresh()
	}

	albumTab := container.NewBorder(
		container.NewVBox(albumSelectArtist, newAlbumEntry, newAlbumBtn, searchAlbumEntry, widget.NewSeparator()),
		nil, nil, nil,
		albumScroll,
	)

	// --- TRACKS ---
	var trackList *widget.List
	trackSelectAlbum := widget.NewSelect(albumNames, nil)
	trackSelectAlbum.PlaceHolder = "Выберите альбом"

	searchTrackEntry := widget.NewEntry()
	searchTrackEntry.SetPlaceHolder("Поиск трека...")
	searchTrackEntry.OnChanged = func(s string) {
		var filteredTracks []Track
		var filteredTrackNames []string
		for _, t := range tracks {
			if s == "" || containsIgnoreCase(t.Title, s) {
				filteredTracks = append(filteredTracks, t)
				filteredTrackNames = append(filteredTrackNames, t.Title)
			}
		}
		tracks = filteredTracks
		trackNames = filteredTrackNames
		trackList.Refresh()
	}

	trackList = widget.NewList(
		func() int { return len(trackNames) },
		func() fyne.CanvasObject { return listRowWithDelete("Трек", func() {}) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(trackNames[i])
			row.Objects[2].(*widget.Button).OnTapped = func() {
				confirmDelete("Удаление трека", "Удалить трек?", func() {
					if err := deleteTrack(tracks[i].ID); err != nil {
						dialog.ShowError(err, mainWindow)
						return
					}
					tracks, trackNames = getTracks()
					trackList.Refresh()
				})
			}
		})

	trackScroll := container.NewVScroll(trackList)
	trackScroll.SetMinSize(fyne.NewSize(600, 200))

	newTrackEntry := widget.NewEntry()
	newTrackEntry.SetPlaceHolder("Название трека")
	newTrackBtn := widget.NewButton("Добавить трек", func() {
		if trackSelectAlbum.Selected == "" || newTrackEntry.Text == "" {
			dialog.ShowInformation("Внимание", "Выберите альбом и введите название трека", mainWindow)
			return
		}
		var albumID int
		for _, al := range albums {
			if al.Title == trackSelectAlbum.Selected {
				albumID = al.ID
				break
			}
		}
		if err := addTrack(newTrackEntry.Text, albumID); err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		tracks, trackNames = getTracks()
		newTrackEntry.SetText("")
		if trackSelectAlbum.Selected != "" {
			trackSelectAlbum.OnChanged(trackSelectAlbum.Selected)
		}
	})

	trackSelectAlbum.OnChanged = func(s string) {
		var selectedAlbumID int
		for _, al := range albums {
			if al.Title == s {
				selectedAlbumID = al.ID
				break
			}
		}
		var filteredTracks []Track
		var filteredTrackNames []string
		for _, t := range tracks {
			if t.AlbumID == selectedAlbumID {
				filteredTracks = append(filteredTracks, t)
				filteredTrackNames = append(filteredTrackNames, t.Title)
			}
		}
		tracks = filteredTracks
		trackNames = filteredTrackNames
		trackList.Refresh()
	}

	trackTab := container.NewBorder(
		container.NewVBox(trackSelectAlbum, newTrackEntry, newTrackBtn, searchTrackEntry, widget.NewSeparator()),
		nil, nil, nil,
		trackScroll,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Артисты", artistTab),
		container.NewTabItem("Альбомы", albumTab),
		container.NewTabItem("Треки", trackTab),
	)

	return container.NewTabItemWithIcon("Моя коллекция", theme.InfoIcon(), tabs)
}

// --- AUTH UI ---
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

// --- MAIN ---
func main() {
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}

	myApp := app.New()
	myApp.Settings().SetTheme(&SpotifyTheme{})

	mainWindow = myApp.NewWindow("Music Manager")
	mainWindow.Resize(fyne.NewSize(800, 700))

	mainWindow.SetContent(createAuthUI(func() {
		mainWindow.SetContent(container.NewAppTabs(
			createPlaylistTab(),
			createDatabaseTab(),
		))
	}))

	mainWindow.ShowAndRun()
}
