//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

func setTrackDuration(t *testing.T, trackID int64, durationMs int) {
	t.Helper()
	_, err := testDB.Exec(context.Background(), `UPDATE tracks SET duration_ms = $1 WHERE id = $2`, durationMs, trackID)
	if err != nil {
		t.Fatalf("set duration on track %d: %v", trackID, err)
	}
}

func setUserSavedAt(t *testing.T, userID, trackID int64, savedAt time.Time) {
	t.Helper()
	_, err := testDB.Exec(context.Background(), `UPDATE user_saved_tracks SET saved_at = $1 WHERE user_id = $2 AND track_id = $3`, savedAt, userID, trackID)
	if err != nil {
		t.Fatalf("set saved_at on user %d track %d: %v", userID, trackID, err)
	}
}

func setArtistMetrics(t *testing.T, artistID int64, popularity, followers int) {
	t.Helper()
	_, err := testDB.Exec(context.Background(), `UPDATE artists SET popularity = $1, followers = $2 WHERE id = $3`, popularity, followers, artistID)
	if err != nil {
		t.Fatalf("set artist metrics on %d: %v", artistID, err)
	}
}

func TestTrackRepo_ListForUser_ReturnsInserted(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_list_user_1")
	artist := insertTestArtist(t, "track_list_artist_1", nil)
	album := insertTestAlbum(t, "track_list_album_1", 2020)
	track := insertTestTrack(t, "track_list_track_1", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkArtist(ctx, track.ID, artist.ID); err != nil {
		t.Fatalf("LinkArtist failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}

	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{})
	if err != nil {
		t.Fatalf("ListForUser failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Items: got %d, want 1", len(result.Items))
	}
	if result.Items[0].SpotifyID != track.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, track.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_SearchFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_search_user_1")
	album := insertTestAlbum(t, "track_search_album_1", 2021)
	track1 := insertTestTrack(t, "track_search_tr_alpha", album.ID, false)
	track2 := insertTestTrack(t, "track_search_tr_beta", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkToUser(ctx, user.ID, track1.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser track1 failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track2.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser track2 failed: %v", err)
	}

	// track1 name is "Track track_search_tr_alpha" — search term matches only it
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "tr_alpha"})
	if err != nil {
		t.Fatalf("ListForUser with search filter failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}

func TestTrackRepo_ListForUser_GenreFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_genre_user_1")
	artist := insertTestArtist(t, "track_genre_artist_1", []string{"rock", "indie"})
	album := insertTestAlbum(t, "track_genre_album_1", 2019)
	track := insertTestTrack(t, "track_genre_track_1", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkArtist(ctx, track.ID, artist.ID); err != nil {
		t.Fatalf("LinkArtist failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}

	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Genres: []string{"rock"}})
	if err != nil {
		t.Fatalf("ListForUser with genre filter failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total with matching genre: got %d, want 1", result.Total)
	}

	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Genres: []string{"jazz"}})
	if err != nil {
		t.Fatalf("ListForUser with non-matching genre failed: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("Total with non-matching genre: got %d, want 0", result.Total)
	}
}

func TestTrackRepo_ListForUser_ExplicitFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_explicit_user_1")
	album := insertTestAlbum(t, "track_explicit_album_1", 2022)
	explicit := insertTestTrack(t, "track_explicit_tr_1", album.ID, true)
	clean := insertTestTrack(t, "track_explicit_tr_2", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkToUser(ctx, user.ID, explicit.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser explicit failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, clean.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser clean failed: %v", err)
	}

	trueVal := true
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Explicit: &trueVal})
	if err != nil {
		t.Fatalf("ListForUser explicit=true failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("explicit=true Total: got %d, want 1", result.Total)
	}

	falseVal := false
	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Explicit: &falseVal})
	if err != nil {
		t.Fatalf("ListForUser explicit=false failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("explicit=false Total: got %d, want 1", result.Total)
	}
}

func TestTrackRepo_ListForUser_SearchFilter_MatchesArtistName(t *testing.T) {
	ctx := context.Background()

	// "artsrch" appears only in the artist name, not in the track or album names,
	// so a match here means the artist predicate is working.
	artist := insertTestArtist(t, "ts_band_artsrch_1", nil)
	album := insertTestAlbum(t, "ts_rec_aaa_1", 2020)
	track := insertTestTrack(t, "ts_song_aaa_1", album.ID, false)

	otherArtist := insertTestArtist(t, "ts_band_bbb_1", nil)
	otherAlbum := insertTestAlbum(t, "ts_rec_bbb_1", 2021)
	otherTrack := insertTestTrack(t, "ts_song_bbb_1", otherAlbum.ID, false)

	user := insertTestUser(t, "ts_user_artsrch_1")
	trackRepo := repository.NewTrackRepo(testDB)

	if err := trackRepo.LinkArtist(ctx, track.ID, artist.ID); err != nil {
		t.Fatalf("LinkArtist failed: %v", err)
	}
	if err := trackRepo.LinkArtist(ctx, otherTrack.ID, otherArtist.ID); err != nil {
		t.Fatalf("LinkArtist other failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, otherTrack.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser other failed: %v", err)
	}

	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "artsrch"})
	if err != nil {
		t.Fatalf("ListForUser with artist name search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != track.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, track.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_SearchFilter_TokensAreAnded(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "ts_token_user_1")
	album := insertTestAlbum(t, "ts_token_album_1", 2020)

	// Names below resolve to "Track <spotifyID>". Track1 contains both
	// "mambo" and "5" but not the contiguous substring "mambo 5".
	track1 := insertTestTrack(t, "mambo_no_5_song", album.ID, false)
	track2 := insertTestTrack(t, "blackmambo_song", album.ID, false)
	track3 := insertTestTrack(t, "year_2025_song", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	for _, tr := range []*models.Track{track1, track2, track3} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}

	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "mambo 5"})
	if err != nil {
		t.Fatalf("ListForUser tokenized search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != track1.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, track1.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_SearchFilter_TrailingSpaceIsNoOp(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "ts_trail_user_1")
	album := insertTestAlbum(t, "ts_trail_album_1", 2020)

	track1 := insertTestTrack(t, "mambo_one_trail", album.ID, false)
	track2 := insertTestTrack(t, "blackmambo_trail", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	for _, tr := range []*models.Track{track1, track2} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}

	plain, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "mambo"})
	if err != nil {
		t.Fatalf("ListForUser plain search failed: %v", err)
	}
	trailing, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "mambo "})
	if err != nil {
		t.Fatalf("ListForUser trailing-space search failed: %v", err)
	}
	if plain.Total != trailing.Total {
		t.Errorf("trailing space changed result: plain Total=%d, trailing Total=%d", plain.Total, trailing.Total)
	}
	if plain.Total != 2 {
		t.Errorf("plain Total: got %d, want 2", plain.Total)
	}
}

func TestTrackRepo_ListForUser_SearchFilter_TokensMatchAcrossFields(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "ts_cross_user_1")
	album := insertTestAlbum(t, "ts_cross_album_1", 2020)
	artist := insertTestArtist(t, "ts_cross_artistxyz_1", nil)
	track := insertTestTrack(t, "ts_cross_song_2020", album.ID, false)

	otherAlbum := insertTestAlbum(t, "ts_cross_album_2", 2020)
	otherTrack := insertTestTrack(t, "ts_cross_song_2020_other", otherAlbum.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkArtist(ctx, track.ID, artist.ID); err != nil {
		t.Fatalf("LinkArtist failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, otherTrack.ID, time.Now()); err != nil {
		t.Fatalf("LinkToUser other failed: %v", err)
	}

	// "artistxyz" is only in the artist name; "2020" is in both track names.
	// AND-ing means only the track linked to that artist matches.
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Search: "artistxyz 2020"})
	if err != nil {
		t.Fatalf("ListForUser cross-field search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != track.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, track.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_DurationFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_duration_user_1")
	album := insertTestAlbum(t, "track_duration_album_1", 2020)
	short := insertTestTrack(t, "track_duration_tr_short", album.ID, false)
	medium := insertTestTrack(t, "track_duration_tr_medium", album.ID, false)
	long := insertTestTrack(t, "track_duration_tr_long", album.ID, false)
	setTrackDuration(t, short.ID, 60_000)
	setTrackDuration(t, medium.ID, 180_000)
	setTrackDuration(t, long.ID, 360_000)

	trackRepo := repository.NewTrackRepo(testDB)
	for _, tr := range []*models.Track{short, medium, long} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}

	min := 120_000
	max := 240_000
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{DurationMin: &min, DurationMax: &max})
	if err != nil {
		t.Fatalf("ListForUser duration range failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != medium.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, medium.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_SavedAtFilter(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_savedat_user_1")
	album := insertTestAlbum(t, "track_savedat_album_1", 2020)
	old := insertTestTrack(t, "track_savedat_tr_old", album.ID, false)
	mid := insertTestTrack(t, "track_savedat_tr_mid", album.ID, false)
	recent := insertTestTrack(t, "track_savedat_tr_recent", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	for _, tr := range []*models.Track{old, mid, recent} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}
	setUserSavedAt(t, user.ID, old.ID, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	setUserSavedAt(t, user.ID, mid.ID, time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
	setUserSavedAt(t, user.ID, recent.ID, time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))

	min := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	max := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{SavedAtMin: &min, SavedAtMax: &max})
	if err != nil {
		t.Fatalf("ListForUser saved_at range failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Items) == 1 && result.Items[0].SpotifyID != mid.SpotifyID {
		t.Errorf("SpotifyID: got %q, want %q", result.Items[0].SpotifyID, mid.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_ArtistPopularityAndFollowersFilters(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_artmetrics_user_1")
	album := insertTestAlbum(t, "track_artmetrics_album_1", 2020)
	smallArtist := insertTestArtist(t, "track_artmetrics_artist_small", nil)
	bigArtist := insertTestArtist(t, "track_artmetrics_artist_big", nil)
	setArtistMetrics(t, smallArtist.ID, 20, 1_000)
	setArtistMetrics(t, bigArtist.ID, 90, 5_000_000)

	trackBySmall := insertTestTrack(t, "track_artmetrics_tr_small", album.ID, false)
	trackByBig := insertTestTrack(t, "track_artmetrics_tr_big", album.ID, false)
	trackRepo := repository.NewTrackRepo(testDB)
	if err := trackRepo.LinkArtist(ctx, trackBySmall.ID, smallArtist.ID); err != nil {
		t.Fatalf("LinkArtist small failed: %v", err)
	}
	if err := trackRepo.LinkArtist(ctx, trackByBig.ID, bigArtist.ID); err != nil {
		t.Fatalf("LinkArtist big failed: %v", err)
	}
	for _, tr := range []*models.Track{trackBySmall, trackByBig} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}

	popMin := 80
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{ArtistPopularityMin: &popMin})
	if err != nil {
		t.Fatalf("ListForUser artist popularity filter failed: %v", err)
	}
	if result.Total != 1 || (len(result.Items) == 1 && result.Items[0].SpotifyID != trackByBig.SpotifyID) {
		t.Errorf("artist_popularity_min=80: got %d items, want 1 (big)", result.Total)
	}

	folMax := 10_000
	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{ArtistFollowersMax: &folMax})
	if err != nil {
		t.Fatalf("ListForUser artist followers filter failed: %v", err)
	}
	if result.Total != 1 || (len(result.Items) == 1 && result.Items[0].SpotifyID != trackBySmall.SpotifyID) {
		t.Errorf("artist_followers_max=10000: got %d items, want 1 (small)", result.Total)
	}
}

func TestTrackRepo_AudioFeatures_UpsertAndFilterAndSort(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_af_user_1")
	album := insertTestAlbum(t, "track_af_album_1", 2020)
	slow := insertTestTrack(t, "track_af_tr_slow", album.ID, false)
	fast := insertTestTrack(t, "track_af_tr_fast", album.ID, false)
	uncategorized := insertTestTrack(t, "track_af_tr_unset", album.ID, false)

	trackRepo := repository.NewTrackRepo(testDB)
	for _, tr := range []*models.Track{slow, fast, uncategorized} {
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser %q failed: %v", tr.SpotifyID, err)
		}
	}

	floatPtr := func(v float64) *float64 { return &v }
	intPtr := func(v int) *int { return &v }

	err := trackRepo.UpsertAudioFeatures(ctx, []repository.AudioFeaturesRow{
		{
			SpotifyID: slow.SpotifyID, Tempo: floatPtr(80), Key: intPtr(0), Mode: intPtr(1),
			TimeSignature: intPtr(4), Energy: floatPtr(0.2), Danceability: floatPtr(0.3),
			Valence: floatPtr(0.4), Acousticness: floatPtr(0.5), Instrumentalness: floatPtr(0.1),
			Liveness: floatPtr(0.1), Speechiness: floatPtr(0.05), Loudness: floatPtr(-12),
		},
		{
			SpotifyID: fast.SpotifyID, Tempo: floatPtr(160), Key: intPtr(7), Mode: intPtr(0),
			TimeSignature: intPtr(3), Energy: floatPtr(0.9), Danceability: floatPtr(0.8),
			Valence: floatPtr(0.7), Acousticness: floatPtr(0.05), Instrumentalness: floatPtr(0.0),
			Liveness: floatPtr(0.4), Speechiness: floatPtr(0.06), Loudness: floatPtr(-4),
		},
	})
	if err != nil {
		t.Fatalf("UpsertAudioFeatures failed: %v", err)
	}

	tempoMin := float64(120)
	result, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{TempoMin: &tempoMin})
	if err != nil {
		t.Fatalf("ListForUser tempo filter failed: %v", err)
	}
	if result.Total != 1 || (len(result.Items) == 1 && result.Items[0].SpotifyID != fast.SpotifyID) {
		t.Errorf("tempo_min=120: got %d items, want 1 (fast)", result.Total)
	}

	majorMode := 1
	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Mode: &majorMode})
	if err != nil {
		t.Fatalf("ListForUser mode filter failed: %v", err)
	}
	if result.Total != 1 || (len(result.Items) == 1 && result.Items[0].SpotifyID != slow.SpotifyID) {
		t.Errorf("mode=1: got %d items, want 1 (slow)", result.Total)
	}

	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Keys: []int{0, 7}})
	if err != nil {
		t.Fatalf("ListForUser key filter failed: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("keys=[0,7]: got %d items, want 2", result.Total)
	}

	result, err = trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{SortBy: "tempo", SortDir: "asc"})
	if err != nil {
		t.Fatalf("ListForUser sort by tempo failed: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Items))
	}
	// NULLS LAST: slow (80), fast (160), uncategorized (NULL)
	if result.Items[0].SpotifyID != slow.SpotifyID ||
		result.Items[1].SpotifyID != fast.SpotifyID ||
		result.Items[2].SpotifyID != uncategorized.SpotifyID {
		t.Errorf("sort tempo asc order wrong: got %q, %q, %q",
			result.Items[0].SpotifyID, result.Items[1].SpotifyID, result.Items[2].SpotifyID)
	}

	missing, err := trackRepo.ListSpotifyIDsMissingAudioFeatures(ctx, user.ID, 30, 100)
	if err != nil {
		t.Fatalf("ListSpotifyIDsMissingAudioFeatures failed: %v", err)
	}
	if len(missing) != 1 || missing[0] != uncategorized.SpotifyID {
		t.Errorf("missing audio features: got %v, want [%q]", missing, uncategorized.SpotifyID)
	}
}

func TestTrackRepo_ListForUser_Pagination(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_page_user_1")
	album := insertTestAlbum(t, "track_page_album_1", 2020)
	trackRepo := repository.NewTrackRepo(testDB)

	for _, sid := range []string{"track_page_tr_a", "track_page_tr_b", "track_page_tr_c"} {
		tr := insertTestTrack(t, sid, album.ID, false)
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID, time.Now()); err != nil {
			t.Fatalf("LinkToUser failed for %q: %v", sid, err)
		}
	}

	page1, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListForUser page 1 failed: %v", err)
	}
	if page1.Total != 3 {
		t.Errorf("Total: got %d, want 3", page1.Total)
	}
	if len(page1.Items) != 2 {
		t.Errorf("page 1 Items: got %d, want 2", len(page1.Items))
	}

	page2, err := trackRepo.ListForUser(ctx, user.ID, models.TrackFilters{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListForUser page 2 failed: %v", err)
	}
	if len(page2.Items) != 1 {
		t.Errorf("page 2 Items: got %d, want 1", len(page2.Items))
	}
}
