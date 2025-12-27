package main

// DATA STRUCTURES
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
	ArtistID int // внешний ключ к таблице artists
	Year     int
}

type Track struct {
	ID       int
	Title    string
	AlbumID  int // Внешний ключ к таблице albums
	Duration int
}
