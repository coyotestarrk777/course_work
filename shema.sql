-- Создание таблицы для пользователей (для аутентификации)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    -- password_hash должен быть достаточно длинным для хранения результата bcrypt (60 символов)
    password_hash VARCHAR(100) NOT NULL, 
    email VARCHAR(100) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для исполнителей
CREATE TABLE IF NOT EXISTS artists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    -- Указываем, что исполнитель создан конкретным пользователем
    user_id INT REFERENCES users(id) ON DELETE SET NULL, 
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для альбомов
CREATE TABLE IF NOT EXISTS albums (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist_id INT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    release_year INT,
    -- Добавим уникальность пары (title, artist_id), чтобы избежать дублирования альбомов у одного исполнителя
    UNIQUE (title, artist_id), 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для треков
CREATE TABLE IF NOT EXISTS tracks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    album_id INT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    duration INT NOT NULL, -- Длительность в секундах
    track_number INT,
    -- Добавим уникальность пары (album_id, track_number) для порядка треков в альбоме
    UNIQUE (album_id, track_number),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_albums_artist_id ON albums(artist_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Уведомление об успешном завершении
SELECT 'Все таблицы успешно созданы или уже существуют.' AS status;