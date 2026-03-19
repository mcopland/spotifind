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
