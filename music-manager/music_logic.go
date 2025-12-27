package main

import (
	"fmt"
)

// --- PLAYLISTS ---

func getPlaylists() ([]Playlist, []string) {
	items, err := repo.GetPlaylists(currentUser.ID)
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, p := range items {
		names = append(names, p.Title)
	}
	return items, names
}

func createPlaylist(title string) error {
	if title == "" {
		return fmt.Errorf("название плейлиста не может быть пустым")
	}
	return repo.CreatePlaylist(title, currentUser.ID)
}

// Получение треков конкретного плейлиста
func getTracksFromPlaylist(playlistID int) ([]Track, []string) {
	items, err := repo.GetTracksFromPlaylist(playlistID)
	if err != nil {
		return nil, nil
	}

	var names []string
	for _, t := range items {
		min := t.Duration / 60
		sec := t.Duration % 60
		names = append(names, fmt.Sprintf("%s (%d:%02d)", t.Title, min, sec))
	}
	return items, names
}

func deletePlaylist(id int) error {
	return repo.DeletePlaylist(id)
}

// --- ARTISTS ---

func getArtists() ([]Artist, []string) {
	items, err := repo.GetArtists()
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, a := range items {
		names = append(names, a.Name)
	}
	return items, names
}

func addArtist(name string) error {
	if name == "" {
		return fmt.Errorf("имя артиста пустое")
	}
	return repo.CreateArtist(name)
}

func deleteArtist(id int) error {
	return repo.DeleteArtist(id)
}

// --- ALBUMS ---

func getAlbums() ([]Album, []string) {
	items, err := repo.GetAlbums()
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, a := range items {
		names = append(names, fmt.Sprintf("%s (%d)", a.Title, a.Year))
	}
	return items, names
}

func addAlbum(title string, artistID, year int) error {
	return repo.CreateAlbum(title, artistID, year)
}

func deleteAlbum(id int) error {
	return repo.DeleteAlbum(id)
}

// --- TRACKS ---

func getTracks() ([]Track, []string) {
	items, err := repo.GetTracks()
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, t := range items {
		min := t.Duration / 60
		sec := t.Duration % 60
		names = append(names, fmt.Sprintf("%s (%d:%02d)", t.Title, min, sec))
	}
	return items, names
}

func addTrack(title string, albumID, duration int) error {
	return repo.CreateTrack(title, albumID, duration)
}

func deleteTrack(id int) error {
	return repo.DeleteTrack(id)
}
