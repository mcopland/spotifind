package spotify

import (
	"context"
	"fmt"
)

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
