CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    nickname TEXT DEFAULT "" ,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    age TEXT NOT NULL,
    gender TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    bio TEXT DEFAULT "",
    isPrivate BOOLEAN NOT NULL DEFAULT 0,
    avatar_url   TEXT    DEFAULT '/img/avatars/images.png'
);