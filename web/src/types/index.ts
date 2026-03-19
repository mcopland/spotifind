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
