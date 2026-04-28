DROP INDEX IF EXISTS idx_user_saved_tracks_saved_at;
DROP INDEX IF EXISTS idx_tracks_duration_ms;
DROP INDEX IF EXISTS idx_tracks_track_key;
DROP INDEX IF EXISTS idx_tracks_danceability;
DROP INDEX IF EXISTS idx_tracks_energy;
DROP INDEX IF EXISTS idx_tracks_tempo;

ALTER TABLE tracks
    DROP COLUMN IF EXISTS audio_features_synced_at,
    DROP COLUMN IF EXISTS loudness,
    DROP COLUMN IF EXISTS speechiness,
    DROP COLUMN IF EXISTS liveness,
    DROP COLUMN IF EXISTS instrumentalness,
    DROP COLUMN IF EXISTS acousticness,
    DROP COLUMN IF EXISTS valence,
    DROP COLUMN IF EXISTS danceability,
    DROP COLUMN IF EXISTS energy,
    DROP COLUMN IF EXISTS time_signature,
    DROP COLUMN IF EXISTS mode,
    DROP COLUMN IF EXISTS track_key,
    DROP COLUMN IF EXISTS tempo;
