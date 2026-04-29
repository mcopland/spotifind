export interface User {
  id: number;
  spotify_id: string;
  display_name: string;
  email: string;
  avatar_url: string;
  last_synced_at: string | null;
}

export interface Artist {
  id?: number;
  spotify_id: string;
  name: string;
  image_url?: string;
  genres?: string[];
  popularity?: number;
  followers?: number;
}

export interface Album {
  id?: number;
  spotify_id: string;
  name: string;
  album_type?: string;
  release_date?: string;
  release_year?: number;
  total_tracks?: number;
  image_url?: string;
  artists?: Artist[];
}

export interface Track {
  id: number;
  spotify_id: string;
  name: string;
  album_id?: number;
  track_number: number;
  duration_ms: number;
  explicit: boolean;
  popularity: number;
  album?: Album;
  artists: Artist[];
  saved_at?: string;
  tempo: number | null;
  key: number | null;
  mode: number | null;
  time_signature: number | null;
  energy: number | null;
  danceability: number | null;
  valence: number | null;
  acousticness: number | null;
  instrumentalness: number | null;
  liveness: number | null;
  speechiness: number | null;
  loudness: number | null;
}

export interface Playlist {
  id: number;
  spotify_id: string;
  name: string;
  description: string;
  owner_id: string;
  is_public: boolean;
  collaborative: boolean;
  image_url?: string;
  track_count?: number;
}

export interface SyncJob {
  id: number;
  user_id: number;
  status: "pending" | "running" | "completed" | "failed" | "none";
  entity_type: string;
  total_items: number;
  synced_items: number;
  error?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
}

export interface PaginatedResult<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface Stats {
  tracks: number;
  albums: number;
  artists: number;
  playlists: number;
}

export interface AlbumTrack {
  spotify_id: string;
  name: string;
  track_number: number;
  duration_ms: number;
  explicit: boolean;
  artists: Artist[];
  liked: boolean;
}

export interface AlbumTracksResponse {
  album: Album | null;
  tracks: AlbumTrack[];
}

export interface TrackFilters {
  search?: string;
  genres?: string[];
  year_min?: number;
  year_max?: number;
  popularity_min?: number;
  popularity_max?: number;
  duration_min?: number;
  duration_max?: number;
  explicit?: boolean;
  playlist?: string;
  saved_at_min?: string;
  saved_at_max?: string;
  artist_popularity_min?: number;
  artist_popularity_max?: number;
  artist_followers_min?: number;
  artist_followers_max?: number;
  tempo_min?: number;
  tempo_max?: number;
  energy_min?: number;
  energy_max?: number;
  danceability_min?: number;
  danceability_max?: number;
  valence_min?: number;
  valence_max?: number;
  acousticness_min?: number;
  acousticness_max?: number;
  instrumentalness_min?: number;
  instrumentalness_max?: number;
  liveness_min?: number;
  liveness_max?: number;
  speechiness_min?: number;
  speechiness_max?: number;
  loudness_min?: number;
  loudness_max?: number;
  keys?: number[];
  mode?: number;
  time_signatures?: number[];
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_dir?: string;
}

export interface AlbumFilters {
  search?: string;
  genres?: string[];
  year_min?: number;
  year_max?: number;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_dir?: string;
}

export interface ArtistFilters {
  search?: string;
  genres?: string[];
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_dir?: string;
}

export type TimeRange = "short_term" | "medium_term" | "long_term";

export interface RecentlyPlayedTrack extends Track {
  played_at: string;
}

export interface TopTrack extends Track {
  rank: number;
  time_range: TimeRange;
}

export interface TopArtist extends Artist {
  rank: number;
  time_range: TimeRange;
}
