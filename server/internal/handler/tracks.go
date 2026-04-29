package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mcopland/spotifind/internal/middleware"
	"github.com/mcopland/spotifind/internal/models"
)

// TrackQuerier is satisfied by repository.TrackRepo.
type TrackQuerier interface {
	ListForUser(ctx context.Context, userID int64, f models.TrackFilters) (*models.PaginatedResult[models.Track], error)
	GetStatsForUser(ctx context.Context, userID int64) (*models.TrackStats, error)
}

type TrackHandler struct {
	trackRepo TrackQuerier
}

func NewTrackHandler(repo TrackQuerier) *TrackHandler {
	return &TrackHandler{trackRepo: repo}
}

func (h *TrackHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	f := parseTrackFilters(r)
	result, err := h.trackRepo.ListForUser(r.Context(), userID, f)
	if err != nil {
		http.Error(w, "failed to list tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *TrackHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.trackRepo.GetStatsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get track stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func parseTrackFilters(r *http.Request) models.TrackFilters {
	q := r.URL.Query()
	f := models.TrackFilters{
		Search:     q.Get("search"),
		PlaylistID: q.Get("playlist"),
		SortBy:     q.Get("sort_by"),
		SortDir:    q.Get("sort_dir"),
	}
	if v := q.Get("artist_id"); v != "" {
		f.ArtistSpotifyID = &v
	}

	if genres := q["genre"]; len(genres) > 0 {
		f.Genres = genres
	}

	parseInt(q, "year_min", &f.YearMin)
	parseInt(q, "year_max", &f.YearMax)
	parseInt(q, "popularity_min", &f.PopularityMin)
	parseInt(q, "popularity_max", &f.PopularityMax)
	parseInt(q, "duration_min", &f.DurationMin)
	parseInt(q, "duration_max", &f.DurationMax)
	parseInt(q, "artist_popularity_min", &f.ArtistPopularityMin)
	parseInt(q, "artist_popularity_max", &f.ArtistPopularityMax)
	parseInt(q, "artist_followers_min", &f.ArtistFollowersMin)
	parseInt(q, "artist_followers_max", &f.ArtistFollowersMax)

	parseFloat(q, "tempo_min", &f.TempoMin)
	parseFloat(q, "tempo_max", &f.TempoMax)
	parseFloat(q, "energy_min", &f.EnergyMin)
	parseFloat(q, "energy_max", &f.EnergyMax)
	parseFloat(q, "danceability_min", &f.DanceabilityMin)
	parseFloat(q, "danceability_max", &f.DanceabilityMax)
	parseFloat(q, "valence_min", &f.ValenceMin)
	parseFloat(q, "valence_max", &f.ValenceMax)
	parseFloat(q, "acousticness_min", &f.AcousticnessMin)
	parseFloat(q, "acousticness_max", &f.AcousticnessMax)
	parseFloat(q, "instrumentalness_min", &f.InstrumentalnessMin)
	parseFloat(q, "instrumentalness_max", &f.InstrumentalnessMax)
	parseFloat(q, "liveness_min", &f.LivenessMin)
	parseFloat(q, "liveness_max", &f.LivenessMax)
	parseFloat(q, "speechiness_min", &f.SpeechinessMin)
	parseFloat(q, "speechiness_max", &f.SpeechinessMax)
	parseFloat(q, "loudness_min", &f.LoudnessMin)
	parseFloat(q, "loudness_max", &f.LoudnessMax)

	parseTime(q, "saved_at_min", &f.SavedAtMin)
	parseTime(q, "saved_at_max", &f.SavedAtMax)

	if v := q.Get("explicit"); v != "" {
		b := v == "true"
		f.Explicit = &b
	}
	if v := q.Get("mode"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Mode = &n
		}
	}
	if keys := q["key"]; len(keys) > 0 {
		f.Keys = parseIntList(keys)
	}
	if sigs := q["time_signature"]; len(sigs) > 0 {
		f.TimeSignatures = parseIntList(sigs)
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Page = n
		}
	}
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.PageSize = n
		}
	}

	return f
}

func parseInt(q url.Values, key string, dst **int) {
	v := q.Get(key)
	if v == "" {
		return
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return
	}
	*dst = &n
}

func parseFloat(q url.Values, key string, dst **float64) {
	v := q.Get(key)
	if v == "" {
		return
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return
	}
	*dst = &n
}

func parseTime(q url.Values, key string, dst **time.Time) {
	v := q.Get(key)
	if v == "" {
		return
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return
	}
	*dst = &t
}

func parseIntList(values []string) []int {
	out := make([]int, 0, len(values))
	for _, v := range values {
		n, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		out = append(out, n)
	}
	return out
}
