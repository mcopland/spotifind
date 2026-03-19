CREATE TABLE IF NOT EXISTS users (
    id                  BIGSERIAL PRIMARY KEY,
    spotify_id          TEXT NOT NULL UNIQUE,
    display_name        TEXT NOT NULL DEFAULT '',
    email               TEXT NOT NULL DEFAULT '',
    avatar_url          TEXT NOT NULL DEFAULT '',
    access_token        TEXT NOT NULL DEFAULT '',
    refresh_token       TEXT NOT NULL DEFAULT '',
    token_expires_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_synced_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS artists (
    id          BIGSERIAL PRIMARY KEY,
    spotify_id  TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    image_url   TEXT NOT NULL DEFAULT '',
    genres      TEXT[] NOT NULL DEFAULT '{}',
    popularity  INT NOT NULL DEFAULT 0,
    followers   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS albums (
    id           BIGSERIAL PRIMARY KEY,
    spotify_id   TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    album_type   TEXT NOT NULL DEFAULT '',
    release_date TEXT NOT NULL DEFAULT '',
    release_year INT NOT NULL DEFAULT 0,
    total_tracks INT NOT NULL DEFAULT 0,
    image_url    TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tracks (
    id           BIGSERIAL PRIMARY KEY,
    spotify_id   TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    album_id     BIGINT REFERENCES albums(id) ON DELETE SET NULL,
    track_number INT NOT NULL DEFAULT 0,
    duration_ms  INT NOT NULL DEFAULT 0,
    explicit     BOOLEAN NOT NULL DEFAULT FALSE,
    popularity   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlists (
    id            BIGSERIAL PRIMARY KEY,
    spotify_id    TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    owner_id      TEXT NOT NULL DEFAULT '',
    is_public     BOOLEAN NOT NULL DEFAULT FALSE,
    collaborative BOOLEAN NOT NULL DEFAULT FALSE,
    snapshot_id   TEXT NOT NULL DEFAULT '',
    image_url     TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS album_artists (
    album_id  BIGINT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    artist_id BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (album_id, artist_id)
);

CREATE TABLE IF NOT EXISTS track_artists (
    track_id  BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    artist_id BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (track_id, artist_id)
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    playlist_id BIGINT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    position    INT NOT NULL DEFAULT 0,
    added_at    TIMESTAMPTZ,
    PRIMARY KEY (playlist_id, track_id)
);

CREATE TABLE IF NOT EXISTS user_saved_tracks (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    saved_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, track_id)
);

CREATE TABLE IF NOT EXISTS user_saved_albums (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   BIGINT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    saved_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, album_id)
);

CREATE TABLE IF NOT EXISTS user_followed_artists (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artist_id  BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    followed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, artist_id)
);

CREATE TABLE IF NOT EXISTS user_playlists (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    playlist_id BIGINT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, playlist_id)
);

CREATE TABLE IF NOT EXISTS user_recently_played (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    played_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, track_id, played_at)
);

CREATE TABLE IF NOT EXISTS user_top_tracks (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   BIGINT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    time_range TEXT NOT NULL,
    rank       INT NOT NULL,
    PRIMARY KEY (user_id, track_id, time_range)
);

CREATE TABLE IF NOT EXISTS user_top_artists (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artist_id  BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    time_range TEXT NOT NULL,
    rank       INT NOT NULL,
    PRIMARY KEY (user_id, artist_id, time_range)
);

CREATE TABLE IF NOT EXISTS sync_jobs (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending',
    entity_type  TEXT NOT NULL DEFAULT 'all',
    total_items  INT NOT NULL DEFAULT 0,
    synced_items INT NOT NULL DEFAULT 0,
    error        TEXT,
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks(album_id);
CREATE INDEX IF NOT EXISTS idx_tracks_popularity ON tracks(popularity);
CREATE INDEX IF NOT EXISTS idx_tracks_name ON tracks USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_albums_release_year ON albums(release_year);
CREATE INDEX IF NOT EXISTS idx_albums_name ON albums USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_artists_name ON artists USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_artists_genres ON artists USING gin(genres);
CREATE INDEX IF NOT EXISTS idx_user_saved_tracks_user_id ON user_saved_tracks(user_id);
CREATE INDEX IF NOT EXISTS idx_user_saved_albums_user_id ON user_saved_albums(user_id);
CREATE INDEX IF NOT EXISTS idx_user_followed_artists_user_id ON user_followed_artists(user_id);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_user_id ON sync_jobs(user_id);
