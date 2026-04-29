import type { Album, AlbumFilters, AlbumTracksResponse, PaginatedResult } from "../types";
import client from "./client";

export async function getAlbums(filters: AlbumFilters = {}): Promise<PaginatedResult<Album>> {
  const params: Record<string, string | string[]> = {};
  if (filters.search) params.search = filters.search;
  if (filters.genres?.length) params.genre = filters.genres;
  if (filters.year_min != null) params.year_min = String(filters.year_min);
  if (filters.year_max != null) params.year_max = String(filters.year_max);
  if (filters.page) params.page = String(filters.page);
  if (filters.page_size) params.page_size = String(filters.page_size);
  if (filters.sort_by) params.sort_by = filters.sort_by;
  if (filters.sort_dir) params.sort_dir = filters.sort_dir;
  const res = await client.get<PaginatedResult<Album>>("/albums", { params });
  return res.data;
}

export async function getAlbumTracks(spotifyID: string): Promise<AlbumTracksResponse> {
  const res = await client.get<AlbumTracksResponse>(`/albums/${spotifyID}/tracks`);
  return res.data;
}
