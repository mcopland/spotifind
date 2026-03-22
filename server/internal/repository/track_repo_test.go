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
