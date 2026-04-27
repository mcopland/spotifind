//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
)

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
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID); err != nil {
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
	if err := trackRepo.LinkToUser(ctx, user.ID, track1.ID); err != nil {
		t.Fatalf("LinkToUser track1 failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, track2.ID); err != nil {
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
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID); err != nil {
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
	if err := trackRepo.LinkToUser(ctx, user.ID, explicit.ID); err != nil {
		t.Fatalf("LinkToUser explicit failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, clean.ID); err != nil {
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
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, otherTrack.ID); err != nil {
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
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID); err != nil {
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
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID); err != nil {
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
	if err := trackRepo.LinkToUser(ctx, user.ID, track.ID); err != nil {
		t.Fatalf("LinkToUser failed: %v", err)
	}
	if err := trackRepo.LinkToUser(ctx, user.ID, otherTrack.ID); err != nil {
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

func TestTrackRepo_ListForUser_Pagination(t *testing.T) {
	ctx := context.Background()
	user := insertTestUser(t, "track_page_user_1")
	album := insertTestAlbum(t, "track_page_album_1", 2020)
	trackRepo := repository.NewTrackRepo(testDB)

	for _, sid := range []string{"track_page_tr_a", "track_page_tr_b", "track_page_tr_c"} {
		tr := insertTestTrack(t, sid, album.ID, false)
		if err := trackRepo.LinkToUser(ctx, user.ID, tr.ID); err != nil {
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
