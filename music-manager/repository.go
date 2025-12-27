package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// AUTH & USERS

func (r *Repository) RegisterUser(u, p string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", u, string(hash))
	return err
}

func (r *Repository) LoginUser(u, p string) (*User, error) {
	var id int
	var hash string
	err := r.db.QueryRow("SELECT id, password_hash FROM users WHERE username=$1", u).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) != nil {
		return nil, fmt.Errorf("неверный логин или пароль")
	}
	return &User{ID: id, Username: u}, nil
}

// ARTISTS

func (r *Repository) GetArtists() ([]Artist, error) {
	rows, err := r.db.Query("SELECT id, name FROM artists WHERE is_deleted=false ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Artist
	for rows.Next() {
		var a Artist
		rows.Scan(&a.ID, &a.Name)
		items = append(items, a)
	}
	return items, nil
}

func (r *Repository) CreateArtist(name string) error {
	_, err := r.db.Exec("INSERT INTO artists (name) VALUES ($1)", name)
	return err
}

func (r *Repository) DeleteArtist(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	// Каскадное удаление
	tx.Exec(`DELETE FROM playlist_tracks WHERE track_id IN (
        SELECT t.id FROM tracks t JOIN albums a ON t.album_id = a.id WHERE a.artist_id = $1
    )`, id)
	tx.Exec("DELETE FROM tracks WHERE album_id IN (SELECT id FROM albums WHERE artist_id = $1)", id)
	tx.Exec("DELETE FROM albums WHERE artist_id = $1", id)
	tx.Exec("DELETE FROM artists WHERE id = $1", id)
	return tx.Commit()
}

// --- ALBUMS ---

func (r *Repository) GetAlbums() ([]Album, error) {
	rows, err := r.db.Query("SELECT id, title, year, artist_id FROM albums WHERE is_deleted=false ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Album
	for rows.Next() {
		var a Album
		rows.Scan(&a.ID, &a.Title, &a.Year, &a.ArtistID)
		items = append(items, a)
	}
	return items, nil
}

func (r *Repository) CreateAlbum(title string, artistID, year int) error {
	_, err := r.db.Exec("INSERT INTO albums (title, artist_id, year) VALUES ($1, $2, $3)", title, artistID, year)
	return err
}

func (r *Repository) DeleteAlbum(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM playlist_tracks WHERE track_id IN (SELECT id FROM tracks WHERE album_id = $1)", id)
	tx.Exec("DELETE FROM tracks WHERE album_id = $1", id)
	tx.Exec("DELETE FROM albums WHERE id = $1", id)
	return tx.Commit()
}

// --- TRACKS ---

func (r *Repository) GetTracks() ([]Track, error) {
	rows, err := r.db.Query("SELECT id, title, album_id, duration FROM tracks WHERE is_deleted=false ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Track
	for rows.Next() {
		var t Track
		rows.Scan(&t.ID, &t.Title, &t.AlbumID, &t.Duration)
		items = append(items, t)
	}
	return items, nil
}

func (r *Repository) CreateTrack(title string, albumID, duration int) error {
	_, err := r.db.Exec("INSERT INTO tracks (title, album_id, duration) VALUES ($1, $2, $3)", title, albumID, duration)
	return err
}

func (r *Repository) DeleteTrack(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM playlist_tracks WHERE track_id = $1", id)
	tx.Exec("DELETE FROM tracks WHERE id = $1", id)
	return tx.Commit()
}

// --- PLAYLISTS ---
func (r *Repository) GetPlaylists(userID int) ([]Playlist, error) {
	rows, err := r.db.Query("SELECT id, title FROM playlists WHERE user_id=$1 AND is_deleted=false ORDER BY title", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Playlist
	for rows.Next() {
		var p Playlist
		rows.Scan(&p.ID, &p.Title)
		items = append(items, p)
	}
	return items, nil
}

func (r *Repository) CreatePlaylist(title string, userID int) error {
	_, err := r.db.Exec("INSERT INTO playlists (title, user_id) VALUES ($1, $2)", title, userID)
	return err
}

func (r *Repository) DeletePlaylist(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM playlist_tracks WHERE playlist_id=$1", id)
	tx.Exec("DELETE FROM playlists WHERE id=$1", id)
	return tx.Commit()
}

func (r *Repository) AddTrackToPlaylist(pID, tID int) error {
	_, err := r.db.Exec("INSERT INTO playlist_tracks (playlist_id, track_id) VALUES ($1, $2)", pID, tID)
	return err
}

func (r *Repository) RemoveTrackFromPlaylist(pID, tID int) error {
	_, err := r.db.Exec("DELETE FROM playlist_tracks WHERE playlist_id=$1 AND track_id=$2", pID, tID)
	return err
}

func (r *Repository) GetTracksFromPlaylist(pID int) ([]Track, error) {
	rows, err := r.db.Query(`
    SELECT t.id, t.title, t.duration 
    FROM tracks t 
    JOIN playlist_tracks pt ON pt.track_id = t.id 
    WHERE pt.playlist_id = $1`, pID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Track
	for rows.Next() {
		var t Track
		rows.Scan(&t.ID, &t.Title, &t.Duration)
		items = append(items, t)
	}
	return items, nil
}
