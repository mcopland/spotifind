ALTER TABLE tracks
    ADD COLUMN tempo                    DOUBLE PRECISION,
    ADD COLUMN track_key                SMALLINT,
    ADD COLUMN mode                     SMALLINT,
    ADD COLUMN time_signature           SMALLINT,
    ADD COLUMN energy                   DOUBLE PRECISION,
    ADD COLUMN danceability             DOUBLE PRECISION,
    ADD COLUMN valence                  DOUBLE PRECISION,
    ADD COLUMN acousticness             DOUBLE PRECISION,
    ADD COLUMN instrumentalness         DOUBLE PRECISION,
    ADD COLUMN liveness                 DOUBLE PRECISION,
    ADD COLUMN speechiness              DOUBLE PRECISION,
    ADD COLUMN loudness                 DOUBLE PRECISION,
    ADD COLUMN audio_features_synced_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tracks_tempo               ON tracks(tempo);
CREATE INDEX IF NOT EXISTS idx_tracks_energy              ON tracks(energy);
CREATE INDEX IF NOT EXISTS idx_tracks_danceability        ON tracks(danceability);
CREATE INDEX IF NOT EXISTS idx_tracks_track_key           ON tracks(track_key);
CREATE INDEX IF NOT EXISTS idx_tracks_duration_ms         ON tracks(duration_ms);
CREATE INDEX IF NOT EXISTS idx_user_saved_tracks_saved_at ON user_saved_tracks(saved_at);
