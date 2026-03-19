import type { PaginatedResult, Playlist, Track, TrackFilters } from "../types";
import client from "./client";

export async function getPlaylists(): Promise<Playlist[]> {
  const res = await client.get<Playlist[]>("/playlists");
  return res.data;
}

export async function getPlaylistTracks(
  id: string,
  filters: TrackFilters = {},
): Promise<PaginatedResult<Track>> {
  const params: Record<string, string | string[]> = {};
  if (filters.search) params.search = filters.search;
  if (filters.page) params.page = String(filters.page);
  if (filters.page_size) params.page_size = String(filters.page_size);
  const res = await client.get<PaginatedResult<Track>>(`/playlists/${id}/tracks`, { params });
  return res.data;
}
