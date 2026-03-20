package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/mcopland/spotifind/internal/models"
	"github.com/mcopland/spotifind/internal/repository"
	"github.com/mcopland/spotifind/internal/spotify"
)

type Service struct {
	trackRepo          *repository.TrackRepo
	albumRepo          *repository.AlbumRepo
	artistRepo         *repository.ArtistRepo
	playlistRepo       *repository.PlaylistRepo
	syncRepo           *repository.SyncRepo
	userRepo           *repository.UserRepo
	authClient         *spotify.AuthClient
	recentlyPlayedRepo *repository.RecentlyPlayedRepo
	topRepo            *repository.TopRepo
}

func NewService(
	trackRepo *repository.TrackRepo,
	albumRepo *repository.AlbumRepo,
	artistRepo *repository.ArtistRepo,
	playlistRepo *repository.PlaylistRepo,
	syncRepo *repository.SyncRepo,
	userRepo *repository.UserRepo,
	authClient *spotify.AuthClient,
	recentlyPlayedRepo *repository.RecentlyPlayedRepo,
	topRepo *repository.TopRepo,
) *Service {
	return &Service{
		trackRepo:          trackRepo,
		albumRepo:          albumRepo,
		artistRepo:         artistRepo,
		playlistRepo:       playlistRepo,
		syncRepo:           syncRepo,
		userRepo:           userRepo,
		authClient:         authClient,
		recentlyPlayedRepo: recentlyPlayedRepo,
		topRepo:            topRepo,
	}
}

func (s *Service) StartSync(userID int64) (int64, error) {
	ctx := context.Background()
	job, err := s.syncRepo.Create(ctx, userID)
	if err != nil {
		return 0, err
	}

	go s.runSync(job.ID, userID)
	return job.ID, nil
}

func (s *Service) runSync(jobID, userID int64) {
	ctx := context.Background()

	if err := s.syncRepo.UpdateStatus(ctx, jobID, "running"); err != nil {
		slog.Error("failed to update sync status", "error", err)
		return
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("get user: %v", err))
		return
	}

	client := spotify.NewClient(
		user.AccessToken, user.RefreshToken, user.TokenExpiresAt,
		s.authClient, userID,
		func(accessToken, refreshToken string, expiresAt time.Time) error {
			return s.userRepo.UpdateTokens(ctx, userID, accessToken, refreshToken, expiresAt)
		},
	)

	if err := s.syncTracks(ctx, client, jobID, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync tracks: %v", err))
		return
	}

	if err := s.syncAlbums(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync albums: %v", err))
		return
	}

	if err := s.syncArtists(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync artists: %v", err))
		return
	}

	if err := s.syncPlaylists(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync playlists: %v", err))
		return
	}

	if err := s.syncRecentlyPlayed(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync recently played: %v", err))
		return
	}

	if err := s.syncTopTracks(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync top tracks: %v", err))
		return
	}

	if err := s.syncTopArtists(ctx, client, userID); err != nil {
		s.failSync(ctx, jobID, fmt.Sprintf("sync top artists: %v", err))
		return
	}

	if err := s.userRepo.UpdateLastSynced(ctx, userID); err != nil {
		slog.Error("failed to update last synced", "error", err)
	}

	if err := s.syncRepo.Finish(ctx, jobID, "completed", nil); err != nil {
		slog.Error("failed to finish sync job", "error", err)
	}
}

func (s *Service) syncTracks(ctx context.Context, client *spotify.Client, jobID, userID int64) error {
	synced := 0
	return client.GetSavedTracks(ctx, func(items []spotify.SavedTrackItem) error {
		for _, item := range items {
			st := item.Track
			if st.ID == "" {
				continue
			}

			al, err := s.upsertAlbum(ctx, st.Album)
			if err != nil {
				return err
			}

			track := &models.Track{
				SpotifyID:   st.ID,
				Name:        st.Name,
				AlbumID:     &al.ID,
				TrackNumber: st.TrackNumber,
				DurationMs:  st.DurationMs,
				Explicit:    st.Explicit,
				Popularity:  st.Popularity,
			}
			saved, err := s.trackRepo.Upsert(ctx, track)
			if err != nil {
				return err
			}

			for _, sa := range st.Artists {
				ar, err := s.upsertArtist(ctx, sa)
				if err != nil {
					return err
				}
				_ = s.trackRepo.LinkArtist(ctx, saved.ID, ar.ID)
			}

			if err := s.trackRepo.LinkToUser(ctx, userID, saved.ID); err != nil {
				return err
			}

			synced++
		}
		return s.syncRepo.UpdateProgress(ctx, jobID, 0, synced)
	})
}

func (s *Service) syncAlbums(ctx context.Context, client *spotify.Client, userID int64) error {
	return client.GetSavedAlbums(ctx, func(items []spotify.SavedAlbumItem) error {
		for _, item := range items {
			al, err := s.upsertAlbum(ctx, item.Album)
			if err != nil {
				return err
			}
			if err := s.albumRepo.LinkToUser(ctx, userID, al.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) syncArtists(ctx context.Context, client *spotify.Client, userID int64) error {
	return client.GetFollowedArtists(ctx, func(items []spotify.SpotifyArtist) error {
		for _, sa := range items {
			ar, err := s.upsertArtist(ctx, sa)
			if err != nil {
				return err
			}
			if err := s.artistRepo.LinkToUser(ctx, userID, ar.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) syncPlaylists(ctx context.Context, client *spotify.Client, userID int64) error {
	return client.GetPlaylists(ctx, func(playlists []spotify.SpotifyPlaylist) error {
		for _, sp := range playlists {
			imageURL := ""
			if len(sp.Images) > 0 {
				imageURL = sp.Images[0].URL
			}
			pl := &models.Playlist{
				SpotifyID:     sp.ID,
				Name:          sp.Name,
				Description:   sp.Description,
				OwnerID:       sp.Owner.ID,
				IsPublic:      sp.Public,
				Collaborative: sp.Collaborative,
				SnapshotID:    sp.SnapshotID,
				ImageURL:      imageURL,
			}
			saved, err := s.playlistRepo.Upsert(ctx, pl)
			if err != nil {
				return err
			}
			if err := s.playlistRepo.LinkToUser(ctx, userID, saved.ID); err != nil {
				return err
			}

			if err := s.playlistRepo.ClearTracks(ctx, saved.ID); err != nil {
				return err
			}

			pos := 0
			if err := client.GetPlaylistTracks(ctx, sp.ID, func(items []spotify.PlaylistTrackItem) error {
				for _, item := range items {
					if item.Track == nil || item.Track.ID == "" {
						pos++
						continue
					}
					al, err := s.upsertAlbum(ctx, item.Track.Album)
					if err != nil {
						return err
					}
					track := &models.Track{
						SpotifyID:   item.Track.ID,
						Name:        item.Track.Name,
						AlbumID:     &al.ID,
						TrackNumber: item.Track.TrackNumber,
						DurationMs:  item.Track.DurationMs,
						Explicit:    item.Track.Explicit,
						Popularity:  item.Track.Popularity,
					}
					savedTrack, err := s.trackRepo.Upsert(ctx, track)
					if err != nil {
						return err
					}
					for _, sa := range item.Track.Artists {
						ar, err := s.upsertArtist(ctx, sa)
						if err != nil {
							return err
						}
						_ = s.trackRepo.LinkArtist(ctx, savedTrack.ID, ar.ID)
					}
					if err := s.playlistRepo.AddTrack(ctx, saved.ID, savedTrack.ID, pos); err != nil {
						return err
					}
					pos++
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) upsertAlbum(ctx context.Context, sa spotify.SpotifyAlbum) (*models.Album, error) {
	imageURL := ""
	if len(sa.Images) > 0 {
		imageURL = sa.Images[0].URL
	}
	year := parseYear(sa.ReleaseDate)
	al := &models.Album{
		SpotifyID:   sa.ID,
		Name:        sa.Name,
		AlbumType:   sa.AlbumType,
		ReleaseDate: sa.ReleaseDate,
		ReleaseYear: year,
		TotalTracks: sa.TotalTracks,
		ImageURL:    imageURL,
	}
	saved, err := s.albumRepo.Upsert(ctx, al)
	if err != nil {
		return nil, err
	}
	for _, sa2 := range sa.Artists {
		ar, err := s.upsertArtist(ctx, sa2)
		if err != nil {
			return nil, err
		}
		_ = s.albumRepo.LinkArtist(ctx, saved.ID, ar.ID)
	}
	return saved, nil
}

func (s *Service) upsertArtist(ctx context.Context, sa spotify.SpotifyArtist) (*models.Artist, error) {
	imageURL := ""
	if len(sa.Images) > 0 {
		imageURL = sa.Images[0].URL
	}
	genres := sa.Genres
	if genres == nil {
		genres = []string{}
	}
	ar := &models.Artist{
		SpotifyID:  sa.ID,
		Name:       sa.Name,
		ImageURL:   imageURL,
		Genres:     genres,
		Popularity: sa.Popularity,
		Followers:  sa.Followers.Total,
	}
	return s.artistRepo.Upsert(ctx, ar)
}

func (s *Service) failSync(ctx context.Context, jobID int64, msg string) {
	slog.Error("sync failed", "job_id", jobID, "error", msg)
	if err := s.syncRepo.Finish(ctx, jobID, "failed", &msg); err != nil {
		slog.Error("failed to record sync failure", "error", err)
	}
}

func parseYear(releaseDate string) int {
	if len(releaseDate) >= 4 {
		y, err := strconv.Atoi(releaseDate[:4])
		if err == nil {
			return y
		}
	}
	return 0
}

func (s *Service) syncRecentlyPlayed(ctx context.Context, client *spotify.Client, userID int64) error {
	items, err := client.GetRecentlyPlayed(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		st := item.Track
		al, err := s.upsertAlbum(ctx, st.Album)
		if err != nil {
			return err
		}
		track := &models.Track{
			SpotifyID:   st.ID,
			Name:        st.Name,
			AlbumID:     &al.ID,
			TrackNumber: st.TrackNumber,
			DurationMs:  st.DurationMs,
			Explicit:    st.Explicit,
			Popularity:  st.Popularity,
		}
		saved, err := s.trackRepo.Upsert(ctx, track)
		if err != nil {
			return err
		}
		for _, sa := range st.Artists {
			ar, err := s.upsertArtist(ctx, sa)
			if err != nil {
				return err
			}
			_ = s.trackRepo.LinkArtist(ctx, saved.ID, ar.ID)
		}
		playedAt, err := time.Parse(time.RFC3339, item.PlayedAt)
		if err != nil {
			return fmt.Errorf("parse played_at %q: %w", item.PlayedAt, err)
		}
		if err := s.recentlyPlayedRepo.Upsert(ctx, userID, saved.ID, playedAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) syncTopTracks(ctx context.Context, client *spotify.Client, userID int64) error {
	for _, timeRange := range []string{"short_term", "medium_term", "long_term"} {
		tracks, err := client.GetTopTracks(ctx, timeRange)
		if err != nil {
			return err
		}
		if err := s.topRepo.DeleteTopTracksForUser(ctx, userID, timeRange); err != nil {
			return err
		}
		for i, st := range tracks {
			al, err := s.upsertAlbum(ctx, st.Album)
			if err != nil {
				return err
			}
			track := &models.Track{
				SpotifyID:   st.ID,
				Name:        st.Name,
				AlbumID:     &al.ID,
				TrackNumber: st.TrackNumber,
				DurationMs:  st.DurationMs,
				Explicit:    st.Explicit,
				Popularity:  st.Popularity,
			}
			saved, err := s.trackRepo.Upsert(ctx, track)
			if err != nil {
				return err
			}
			for _, sa := range st.Artists {
				ar, err := s.upsertArtist(ctx, sa)
				if err != nil {
					return err
				}
				_ = s.trackRepo.LinkArtist(ctx, saved.ID, ar.ID)
			}
			if err := s.topRepo.UpsertTopTrack(ctx, userID, saved.ID, i+1, timeRange); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) syncTopArtists(ctx context.Context, client *spotify.Client, userID int64) error {
	for _, timeRange := range []string{"short_term", "medium_term", "long_term"} {
		artists, err := client.GetTopArtists(ctx, timeRange)
		if err != nil {
			return err
		}
		if err := s.topRepo.DeleteTopArtistsForUser(ctx, userID, timeRange); err != nil {
			return err
		}
		for i, sa := range artists {
			ar, err := s.upsertArtist(ctx, sa)
			if err != nil {
				return err
			}
			if err := s.topRepo.UpsertTopArtist(ctx, userID, ar.ID, i+1, timeRange); err != nil {
				return err
			}
		}
	}
	return nil
}
