package spotify

import (
	"context"
	"fmt"
	"strings"
)

type recentlyPlayedResponse struct {
	Items []PlayHistoryItem `json:"items"`
}

type PlayHistoryItem struct {
	PlayedAt string        `json:"played_at"`
	Track    *SpotifyTrack `json:"track"`
}

type SpotifyUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Images      []struct {
		URL string `json:"url"`
	} `json:"images"`
}

type SpotifyImage struct {
	URL string `json:"url"`
}

type SpotifyArtist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images []SpotifyImage `json:"images"`
}

type SpotifyAlbum struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	AlbumType   string          `json:"album_type"`
	ReleaseDate string          `json:"release_date"`
	TotalTracks int             `json:"total_tracks"`
	Images      []SpotifyImage  `json:"images"`
	Artists     []SpotifyArtist `json:"artists"`
}

type SpotifyTrack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TrackNumber int             `json:"track_number"`
	DurationMs  int             `json:"duration_ms"`
	Explicit    bool            `json:"explicit"`
	Popularity  int             `json:"popularity"`
	Album       SpotifyAlbum    `json:"album"`
	Artists     []SpotifyArtist `json:"artists"`
}

type SavedTrackItem struct {
	AddedAt string       `json:"added_at"`
	Track   SpotifyTrack `json:"track"`
}

type SavedAlbumItem struct {
	AddedAt string       `json:"added_at"`
	Album   SpotifyAlbum `json:"album"`
}

type SpotifyPlaylist struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Public        bool           `json:"public"`
	Collaborative bool           `json:"collaborative"`
	SnapshotID    string         `json:"snapshot_id"`
	Images        []SpotifyImage `json:"images"`
	Owner         struct {
		ID string `json:"id"`
	} `json:"owner"`
	Tracks struct {
		Total int `json:"total"`
	} `json:"tracks"`
}

type PlaylistTrackItem struct {
	AddedAt string        `json:"added_at"`
	Track   *SpotifyTrack `json:"track"`
}

type pagedResponse[T any] struct {
	Items  []T    `json:"items"`
	Next   string `json:"next"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type followedArtistsResponse struct {
	Artists struct {
		Items   []SpotifyArtist `json:"items"`
		Next    string          `json:"next"`
		Cursors struct {
			After string `json:"after"`
		} `json:"cursors"`
	} `json:"artists"`
}

func (c *Client) GetCurrentUser(ctx context.Context) (*SpotifyUser, error) {
	var u SpotifyUser
	if err := c.Get(ctx, "/me", &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *Client) GetSavedTracks(ctx context.Context, onBatch func([]SavedTrackItem) error) error {
	offset := 0
	limit := 50
	for {
		var page pagedResponse[SavedTrackItem]
		path := fmt.Sprintf("/me/tracks?limit=%d&offset=%d", limit, offset)
		if err := c.Get(ctx, path, &page); err != nil {
			return err
		}
		if len(page.Items) == 0 {
			break
		}
		if err := onBatch(page.Items); err != nil {
			return err
		}
		offset += len(page.Items)
		if page.Next == "" {
			break
		}
	}
	return nil
}

func (c *Client) GetSavedAlbums(ctx context.Context, onBatch func([]SavedAlbumItem) error) error {
	offset := 0
	limit := 50
	for {
		var page pagedResponse[SavedAlbumItem]
		path := fmt.Sprintf("/me/albums?limit=%d&offset=%d", limit, offset)
		if err := c.Get(ctx, path, &page); err != nil {
			return err
		}
		if len(page.Items) == 0 {
			break
		}
		if err := onBatch(page.Items); err != nil {
			return err
		}
		offset += len(page.Items)
		if page.Next == "" {
			break
		}
	}
	return nil
}

func (c *Client) GetFollowedArtists(ctx context.Context, onBatch func([]SpotifyArtist) error) error {
	after := ""
	limit := 50
	for {
		path := fmt.Sprintf("/me/following?type=artist&limit=%d", limit)
		if after != "" {
			path += "&after=" + after
		}
		var resp followedArtistsResponse
		if err := c.Get(ctx, path, &resp); err != nil {
			return err
		}
		items := resp.Artists.Items
		if len(items) == 0 {
			break
		}
		if err := onBatch(items); err != nil {
			return err
		}
		after = resp.Artists.Cursors.After
		if resp.Artists.Next == "" {
			break
		}
	}
	return nil
}

func (c *Client) GetPlaylists(ctx context.Context, onBatch func([]SpotifyPlaylist) error) error {
	offset := 0
	limit := 50
	for {
		var page pagedResponse[SpotifyPlaylist]
		path := fmt.Sprintf("/me/playlists?limit=%d&offset=%d", limit, offset)
		if err := c.Get(ctx, path, &page); err != nil {
			return err
		}
		if len(page.Items) == 0 {
			break
		}
		if err := onBatch(page.Items); err != nil {
			return err
		}
		offset += len(page.Items)
		if page.Next == "" {
			break
		}
	}
	return nil
}

func (c *Client) GetPlaylistTracks(ctx context.Context, playlistID string, onBatch func([]PlaylistTrackItem) error) error {
	offset := 0
	limit := 100
	for {
		var page pagedResponse[PlaylistTrackItem]
		path := fmt.Sprintf("/playlists/%s/tracks?limit=%d&offset=%d", playlistID, limit, offset)
		if err := c.Get(ctx, path, &page); err != nil {
			return err
		}
		if len(page.Items) == 0 {
			break
		}
		if err := onBatch(page.Items); err != nil {
			return err
		}
		offset += len(page.Items)
		if page.Next == "" {
			break
		}
	}
	return nil
}

// GetRecentlyPlayed returns up to 50 recently played tracks.
// Items with a nil or empty-ID track are omitted.
func (c *Client) GetRecentlyPlayed(ctx context.Context) ([]PlayHistoryItem, error) {
	var resp recentlyPlayedResponse
	if err := c.Get(ctx, "/me/player/recently-played?limit=50", &resp); err != nil {
		return nil, fmt.Errorf("get recently played: %w", err)
	}
	out := make([]PlayHistoryItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.Track == nil || item.Track.ID == "" {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// GetTopTracks returns up to 50 top tracks for the given time range.
func (c *Client) GetTopTracks(ctx context.Context, timeRange string) ([]SpotifyTrack, error) {
	var resp pagedResponse[SpotifyTrack]
	path := fmt.Sprintf("/me/top/tracks?time_range=%s&limit=50", timeRange)
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get top tracks %s: %w", timeRange, err)
	}
	return resp.Items, nil
}

// GetTopArtists returns up to 50 top artists for the given time range.
func (c *Client) GetTopArtists(ctx context.Context, timeRange string) ([]SpotifyArtist, error) {
	var resp pagedResponse[SpotifyArtist]
	path := fmt.Sprintf("/me/top/artists?time_range=%s&limit=50", timeRange)
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get top artists %s: %w", timeRange, err)
	}
	return resp.Items, nil
}

const audioFeaturesBatchLimit = 100

type AudioFeatures struct {
	ID               string  `json:"id"`
	Tempo            float64 `json:"tempo"`
	Key              int     `json:"key"`
	Mode             int     `json:"mode"`
	TimeSignature    int     `json:"time_signature"`
	Energy           float64 `json:"energy"`
	Danceability     float64 `json:"danceability"`
	Valence          float64 `json:"valence"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Speechiness      float64 `json:"speechiness"`
	Loudness         float64 `json:"loudness"`
}

type audioFeaturesResponse struct {
	AudioFeatures []*AudioFeatures `json:"audio_features"`
}

// GetAudioFeaturesBatch fetches features for up to 100 track IDs per call.
// Spotify returns nil entries for IDs it cannot resolve; nils are preserved
// in the result slice so callers can correlate by index if needed.
func (c *Client) GetAudioFeaturesBatch(ctx context.Context, ids []string) ([]*AudioFeatures, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > audioFeaturesBatchLimit {
		return nil, fmt.Errorf("get audio features: batch size %d exceeds limit %d", len(ids), audioFeaturesBatchLimit)
	}
	var resp audioFeaturesResponse
	path := fmt.Sprintf("/audio-features?ids=%s", strings.Join(ids, ","))
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get audio features for %d ids: %w", len(ids), err)
	}
	return resp.AudioFeatures, nil
}
