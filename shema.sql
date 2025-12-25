-- ================= USERS =================
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

-- ================= ARTISTS =================
CREATE TABLE artists (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    is_deleted BOOLEAN DEFAULT false
);

CREATE UNIQUE INDEX artists_name_unique
ON artists (name)
WHERE is_deleted = false;

-- ================= ALBUMS =================
CREATE TABLE albums (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    year INTEGER,
    artist_id INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    is_deleted BOOLEAN DEFAULT false
);

CREATE UNIQUE INDEX albums_artist_title_unique
ON albums (artist_id, title)
WHERE is_deleted = false;

-- ================= TRACKS =================
CREATE TABLE tracks (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    album_id INTEGER NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    duration INTEGER,
    is_deleted BOOLEAN DEFAULT false
);

CREATE UNIQUE INDEX tracks_album_title_unique
ON tracks (album_id, title)
WHERE is_deleted = false;

-- ================= PLAYLISTS =================
CREATE TABLE playlists (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_deleted BOOLEAN DEFAULT false
);

CREATE UNIQUE INDEX playlists_user_title_unique
ON playlists (user_id, title)
WHERE is_deleted = false;

-- ================= PLAYLIST_TRACKS =================
CREATE TABLE playlist_tracks (
    playlist_id INTEGER REFERENCES playlists(id) ON DELETE CASCADE,
    track_id INTEGER REFERENCES tracks(id) ON DELETE CASCADE,
    PRIMARY KEY (playlist_id, track_id)
);
