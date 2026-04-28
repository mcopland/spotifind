import type { PaginatedResult, Track, TrackFilters } from "../types";
import client from "./client";

function buildParams(f: TrackFilters): Record<string, string | string[]> {
  const params: Record<string, string | string[]> = {};
  if (f.search) params.search = f.search;
  if (f.genres?.length) params.genre = f.genres;
  if (f.year_min != null) params.year_min = String(f.year_min);
  if (f.year_max != null) params.year_max = String(f.year_max);
  if (f.popularity_min != null) params.popularity_min = String(f.popularity_min);
  if (f.popularity_max != null) params.popularity_max = String(f.popularity_max);
  if (f.duration_min != null) params.duration_min = String(f.duration_min);
  if (f.duration_max != null) params.duration_max = String(f.duration_max);
  if (f.explicit != null) params.explicit = String(f.explicit);
  if (f.playlist) params.playlist = f.playlist;
  if (f.saved_at_min) params.saved_at_min = f.saved_at_min;
  if (f.saved_at_max) params.saved_at_max = f.saved_at_max;
  if (f.artist_popularity_min != null) params.artist_popularity_min = String(f.artist_popularity_min);
  if (f.artist_popularity_max != null) params.artist_popularity_max = String(f.artist_popularity_max);
  if (f.artist_followers_min != null) params.artist_followers_min = String(f.artist_followers_min);
  if (f.artist_followers_max != null) params.artist_followers_max = String(f.artist_followers_max);
  if (f.tempo_min != null) params.tempo_min = String(f.tempo_min);
  if (f.tempo_max != null) params.tempo_max = String(f.tempo_max);
  if (f.energy_min != null) params.energy_min = String(f.energy_min);
  if (f.energy_max != null) params.energy_max = String(f.energy_max);
  if (f.danceability_min != null) params.danceability_min = String(f.danceability_min);
  if (f.danceability_max != null) params.danceability_max = String(f.danceability_max);
  if (f.valence_min != null) params.valence_min = String(f.valence_min);
  if (f.valence_max != null) params.valence_max = String(f.valence_max);
  if (f.acousticness_min != null) params.acousticness_min = String(f.acousticness_min);
  if (f.acousticness_max != null) params.acousticness_max = String(f.acousticness_max);
  if (f.instrumentalness_min != null) params.instrumentalness_min = String(f.instrumentalness_min);
  if (f.instrumentalness_max != null) params.instrumentalness_max = String(f.instrumentalness_max);
  if (f.liveness_min != null) params.liveness_min = String(f.liveness_min);
  if (f.liveness_max != null) params.liveness_max = String(f.liveness_max);
  if (f.speechiness_min != null) params.speechiness_min = String(f.speechiness_min);
  if (f.speechiness_max != null) params.speechiness_max = String(f.speechiness_max);
  if (f.loudness_min != null) params.loudness_min = String(f.loudness_min);
  if (f.loudness_max != null) params.loudness_max = String(f.loudness_max);
  if (f.keys?.length) params.key = f.keys.map(String);
  if (f.mode != null) params.mode = String(f.mode);
  if (f.time_signatures?.length) params.time_signature = f.time_signatures.map(String);
  if (f.page) params.page = String(f.page);
  if (f.page_size) params.page_size = String(f.page_size);
  if (f.sort_by) params.sort_by = f.sort_by;
  if (f.sort_dir) params.sort_dir = f.sort_dir;
  return params;
}

export async function getTracks(filters: TrackFilters = {}): Promise<PaginatedResult<Track>> {
  const res = await client.get<PaginatedResult<Track>>("/tracks", { params: buildParams(filters) });
  return res.data;
}
