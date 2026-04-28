package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mcopland/spotifind/internal/models"
)

type TrackRepo struct {
	db *pgxpool.Pool
}

func NewTrackRepo(db *pgxpool.Pool) *TrackRepo {
	return &TrackRepo{db: db}
}

func (r *TrackRepo) Upsert(ctx context.Context, t *models.Track) (*models.Track, error) {
	query := `
		INSERT INTO tracks (spotify_id, name, album_id, track_number, duration_ms, explicit, popularity, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (spotify_id) DO UPDATE SET
			name         = EXCLUDED.name,
			album_id     = EXCLUDED.album_id,
			track_number = EXCLUDED.track_number,
			duration_ms  = EXCLUDED.duration_ms,
			explicit     = EXCLUDED.explicit,
			popularity   = EXCLUDED.popularity,
			updated_at   = NOW()
		RETURNING id, spotify_id, name, album_id, track_number, duration_ms, explicit, popularity, created_at, updated_at`

	tr := &models.Track{}
	err := r.db.QueryRow(ctx, query, t.SpotifyID, t.Name, t.AlbumID, t.TrackNumber, t.DurationMs, t.Explicit, t.Popularity).
		Scan(&tr.ID, &tr.SpotifyID, &tr.Name, &tr.AlbumID, &tr.TrackNumber, &tr.DurationMs, &tr.Explicit, &tr.Popularity, &tr.CreatedAt, &tr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert track %s: %w", t.SpotifyID, err)
	}
	return tr, nil
}

func (r *TrackRepo) LinkArtist(ctx context.Context, trackID, artistID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO track_artists (track_id, artist_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		trackID, artistID)
	if err != nil {
		return fmt.Errorf("link track %d to artist %d: %w", trackID, artistID, err)
	}
	return nil
}

func (r *TrackRepo) LinkToUser(ctx context.Context, userID, trackID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_saved_tracks (user_id, track_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, trackID)
	if err != nil {
		return fmt.Errorf("link user %d to track %d: %w", userID, trackID, err)
	}
	return nil
}

func (r *TrackRepo) ListForUser(ctx context.Context, userID int64, f models.TrackFilters) (*models.PaginatedResult[models.Track], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	args := []any{userID}
	where := []string{"ust.user_id = $1"}
	idx := 2

	for _, token := range strings.Fields(f.Search) {
		where = append(where, fmt.Sprintf(`(
			t.name ILIKE $%d
			OR al.name ILIKE $%d
			OR EXISTS (
				SELECT 1 FROM track_artists ta_s
				JOIN artists ar_s ON ar_s.id = ta_s.artist_id
				WHERE ta_s.track_id = t.id AND ar_s.name ILIKE $%d
			)
		)`, idx, idx, idx))
		args = append(args, "%"+token+"%")
		idx++
	}
	if len(f.Genres) > 0 {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM track_artists ta2
			JOIN artists ar2 ON ar2.id = ta2.artist_id
			WHERE ta2.track_id = t.id AND ar2.genres && $%d)`, idx))
		args = append(args, f.Genres)
		idx++
	}
	if f.YearMin != nil {
		where = append(where, fmt.Sprintf("al.release_year >= $%d", idx))
		args = append(args, *f.YearMin)
		idx++
	}
	if f.YearMax != nil {
		where = append(where, fmt.Sprintf("al.release_year <= $%d", idx))
		args = append(args, *f.YearMax)
		idx++
	}
	if f.PopularityMin != nil {
		where = append(where, fmt.Sprintf("t.popularity >= $%d", idx))
		args = append(args, *f.PopularityMin)
		idx++
	}
	if f.PopularityMax != nil {
		where = append(where, fmt.Sprintf("t.popularity <= $%d", idx))
		args = append(args, *f.PopularityMax)
		idx++
	}
	if f.DurationMin != nil {
		where = append(where, fmt.Sprintf("t.duration_ms >= $%d", idx))
		args = append(args, *f.DurationMin)
		idx++
	}
	if f.DurationMax != nil {
		where = append(where, fmt.Sprintf("t.duration_ms <= $%d", idx))
		args = append(args, *f.DurationMax)
		idx++
	}
	if f.Explicit != nil {
		where = append(where, fmt.Sprintf("t.explicit = $%d", idx))
		args = append(args, *f.Explicit)
		idx++
	}
	if f.PlaylistID != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM playlist_tracks pt
			JOIN playlists p ON p.id = pt.playlist_id
			WHERE pt.track_id = t.id AND p.spotify_id = $%d)`, idx))
		args = append(args, f.PlaylistID)
		idx++
	}
	if f.SavedAtMin != nil {
		where = append(where, fmt.Sprintf("ust.saved_at >= $%d", idx))
		args = append(args, *f.SavedAtMin)
		idx++
	}
	if f.SavedAtMax != nil {
		where = append(where, fmt.Sprintf("ust.saved_at <= $%d", idx))
		args = append(args, *f.SavedAtMax)
		idx++
	}
	if f.ArtistPopularityMin != nil || f.ArtistPopularityMax != nil ||
		f.ArtistFollowersMin != nil || f.ArtistFollowersMax != nil {
		clauses := []string{}
		if f.ArtistPopularityMin != nil {
			clauses = append(clauses, fmt.Sprintf("ar3.popularity >= $%d", idx))
			args = append(args, *f.ArtistPopularityMin)
			idx++
		}
		if f.ArtistPopularityMax != nil {
			clauses = append(clauses, fmt.Sprintf("ar3.popularity <= $%d", idx))
			args = append(args, *f.ArtistPopularityMax)
			idx++
		}
		if f.ArtistFollowersMin != nil {
			clauses = append(clauses, fmt.Sprintf("ar3.followers >= $%d", idx))
			args = append(args, *f.ArtistFollowersMin)
			idx++
		}
		if f.ArtistFollowersMax != nil {
			clauses = append(clauses, fmt.Sprintf("ar3.followers <= $%d", idx))
			args = append(args, *f.ArtistFollowersMax)
			idx++
		}
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM track_artists ta3
			JOIN artists ar3 ON ar3.id = ta3.artist_id
			WHERE ta3.track_id = t.id AND %s)`, strings.Join(clauses, " AND ")))
	}

	floatRange := func(col string, min, max *float64) {
		if min != nil {
			where = append(where, fmt.Sprintf("%s >= $%d", col, idx))
			args = append(args, *min)
			idx++
		}
		if max != nil {
			where = append(where, fmt.Sprintf("%s <= $%d", col, idx))
			args = append(args, *max)
			idx++
		}
	}
	floatRange("t.tempo", f.TempoMin, f.TempoMax)
	floatRange("t.energy", f.EnergyMin, f.EnergyMax)
	floatRange("t.danceability", f.DanceabilityMin, f.DanceabilityMax)
	floatRange("t.valence", f.ValenceMin, f.ValenceMax)
	floatRange("t.acousticness", f.AcousticnessMin, f.AcousticnessMax)
	floatRange("t.instrumentalness", f.InstrumentalnessMin, f.InstrumentalnessMax)
	floatRange("t.liveness", f.LivenessMin, f.LivenessMax)
	floatRange("t.speechiness", f.SpeechinessMin, f.SpeechinessMax)
	floatRange("t.loudness", f.LoudnessMin, f.LoudnessMax)

	if len(f.Keys) > 0 {
		where = append(where, fmt.Sprintf("t.track_key = ANY($%d)", idx))
		args = append(args, f.Keys)
		idx++
	}
	if f.Mode != nil {
		where = append(where, fmt.Sprintf("t.mode = $%d", idx))
		args = append(args, *f.Mode)
		idx++
	}
	if len(f.TimeSignatures) > 0 {
		where = append(where, fmt.Sprintf("t.time_signature = ANY($%d)", idx))
		args = append(args, f.TimeSignatures)
		idx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT t.id)
		FROM tracks t
		JOIN user_saved_tracks ust ON t.id = ust.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		%s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count tracks for user %d: %w", userID, err)
	}

	sortCol := "t.name"
	switch f.SortBy {
	case "popularity":
		sortCol = "t.popularity"
	case "duration":
		sortCol = "t.duration_ms"
	case "year":
		sortCol = "al.release_year"
	case "album":
		sortCol = "al.name"
	case "saved_at":
		sortCol = "ust.saved_at"
	case "tempo":
		sortCol = "t.tempo"
	case "energy":
		sortCol = "t.energy"
	case "danceability":
		sortCol = "t.danceability"
	case "artist_popularity":
		sortCol = "(SELECT MAX(ar4.popularity) FROM track_artists ta4 JOIN artists ar4 ON ar4.id = ta4.artist_id WHERE ta4.track_id = t.id)"
	case "artist_followers":
		sortCol = "(SELECT MAX(ar5.followers) FROM track_artists ta5 JOIN artists ar5 ON ar5.id = ta5.artist_id WHERE ta5.track_id = t.id)"
	}
	sortDir := "ASC"
	if strings.ToUpper(f.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	offset := (f.Page - 1) * f.PageSize
	listQuery := fmt.Sprintf(`
		SELECT t.id, t.spotify_id, t.name, t.album_id, t.track_number, t.duration_ms, t.explicit, t.popularity,
		       t.tempo, t.track_key, t.mode, t.time_signature,
		       t.energy, t.danceability, t.valence, t.acousticness,
		       t.instrumentalness, t.liveness, t.speechiness, t.loudness,
		       t.audio_features_synced_at,
		       t.created_at, t.updated_at, ust.saved_at,
		       al.id, al.spotify_id, al.name, al.album_type, al.release_date, al.release_year, al.total_tracks, al.image_url, al.created_at, al.updated_at,
		       COALESCE(array_agg(DISTINCT ar.name ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_names,
		       COALESCE(array_agg(ar.spotify_id ORDER BY ar.name) FILTER (WHERE ar.id IS NOT NULL), '{}') AS artist_spotify_ids
		FROM tracks t
		JOIN user_saved_tracks ust ON t.id = ust.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		LEFT JOIN track_artists ta ON ta.track_id = t.id
		LEFT JOIN artists ar ON ar.id = ta.artist_id
		%s
		GROUP BY t.id, ust.saved_at, al.id
		ORDER BY %s %s NULLS LAST, t.id ASC
		LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, idx, idx+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list tracks for user %d: %w", userID, err)
	}
	defer rows.Close()

	tracks := make([]models.Track, 0)
	for rows.Next() {
		tr := &models.Track{}
		al := &models.Album{}
		var artistNames, artistSpotifyIDs []string
		err := rows.Scan(
			&tr.ID, &tr.SpotifyID, &tr.Name, &tr.AlbumID, &tr.TrackNumber, &tr.DurationMs, &tr.Explicit, &tr.Popularity,
			&tr.Tempo, &tr.Key, &tr.Mode, &tr.TimeSignature,
			&tr.Energy, &tr.Danceability, &tr.Valence, &tr.Acousticness,
			&tr.Instrumentalness, &tr.Liveness, &tr.Speechiness, &tr.Loudness,
			&tr.AudioFeaturesSyncedAt,
			&tr.CreatedAt, &tr.UpdatedAt, &tr.SavedAt,
			&al.ID, &al.SpotifyID, &al.Name, &al.AlbumType, &al.ReleaseDate, &al.ReleaseYear, &al.TotalTracks, &al.ImageURL, &al.CreatedAt, &al.UpdatedAt,
			&artistNames, &artistSpotifyIDs,
		)
		if err != nil {
			return nil, fmt.Errorf("scan track row: %w", err)
		}
		tr.Album = al
		for i, name := range artistNames {
			spotifyID := ""
			if i < len(artistSpotifyIDs) {
				spotifyID = artistSpotifyIDs[i]
			}
			tr.Artists = append(tr.Artists, models.Artist{Name: name, SpotifyID: spotifyID})
		}
		tracks = append(tracks, *tr)
	}

	return &models.PaginatedResult[models.Track]{
		Items:    tracks,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

// AudioFeaturesRow carries the per-track values written by UpsertAudioFeatures.
// Pointer fields permit NULL writes for unknown values (e.g., Spotify key = -1).
type AudioFeaturesRow struct {
	SpotifyID        string
	Tempo            *float64
	Key              *int
	Mode             *int
	TimeSignature    *int
	Energy           *float64
	Danceability     *float64
	Valence          *float64
	Acousticness     *float64
	Instrumentalness *float64
	Liveness         *float64
	Speechiness      *float64
	Loudness         *float64
}

// ListSpotifyIDsMissingAudioFeatures returns spotify IDs of tracks linked to
// the user where audio features have never been pulled or are stale beyond
// staleAfterDays. Caller bounds the result with limit.
func (r *TrackRepo) ListSpotifyIDsMissingAudioFeatures(ctx context.Context, userID int64, staleAfterDays, limit int) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.spotify_id
		FROM tracks t
		JOIN user_saved_tracks ust ON ust.track_id = t.id
		WHERE ust.user_id = $1
		  AND (t.audio_features_synced_at IS NULL
		       OR t.audio_features_synced_at < NOW() - ($2::int || ' days')::interval)
		ORDER BY t.audio_features_synced_at NULLS FIRST, t.id ASC
		LIMIT $3`, userID, staleAfterDays, limit)
	if err != nil {
		return nil, fmt.Errorf("list tracks missing audio features for user %d: %w", userID, err)
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan spotify id: %w", err)
		}
		out = append(out, id)
	}
	return out, nil
}

// UpsertAudioFeatures bulk-updates a batch of tracks by spotify_id.
// Each call issues a single UPDATE ... FROM (VALUES ...) statement.
// Tracks not present in the tracks table are silently skipped.
func (r *TrackRepo) UpsertAudioFeatures(ctx context.Context, batch []AudioFeaturesRow) error {
	if len(batch) == 0 {
		return nil
	}
	const colsPerRow = 13
	args := make([]any, 0, len(batch)*colsPerRow)
	values := make([]string, 0, len(batch))
	for i, row := range batch {
		base := i * colsPerRow
		values = append(values, fmt.Sprintf(
			"($%d, $%d::double precision, $%d::smallint, $%d::smallint, $%d::smallint, $%d::double precision, $%d::double precision, $%d::double precision, $%d::double precision, $%d::double precision, $%d::double precision, $%d::double precision, $%d::double precision)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12, base+13,
		))
		args = append(args,
			row.SpotifyID, row.Tempo, row.Key, row.Mode, row.TimeSignature,
			row.Energy, row.Danceability, row.Valence, row.Acousticness,
			row.Instrumentalness, row.Liveness, row.Speechiness, row.Loudness,
		)
	}
	query := fmt.Sprintf(`
		UPDATE tracks AS t SET
			tempo                    = v.tempo,
			track_key                = v.track_key,
			mode                     = v.mode,
			time_signature           = v.time_signature,
			energy                   = v.energy,
			danceability             = v.danceability,
			valence                  = v.valence,
			acousticness             = v.acousticness,
			instrumentalness         = v.instrumentalness,
			liveness                 = v.liveness,
			speechiness              = v.speechiness,
			loudness                 = v.loudness,
			audio_features_synced_at = NOW(),
			updated_at               = NOW()
		FROM (VALUES %s) AS v(spotify_id, tempo, track_key, mode, time_signature, energy, danceability, valence, acousticness, instrumentalness, liveness, speechiness, loudness)
		WHERE t.spotify_id = v.spotify_id`, strings.Join(values, ", "))
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("upsert audio features for %d tracks: %w", len(batch), err)
	}
	return nil
}
