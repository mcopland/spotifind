import type { Artist, ArtistFilters, PaginatedResult } from "../types";
import client from "./client";

export async function getArtists(filters: ArtistFilters = {}): Promise<PaginatedResult<Artist>> {
  const params: Record<string, string | string[]> = {};
  if (filters.search) params.search = filters.search;
  if (filters.genres?.length) params.genre = filters.genres;
  if (filters.page) params.page = String(filters.page);
  if (filters.page_size) params.page_size = String(filters.page_size);
  if (filters.sort_by) params.sort_by = filters.sort_by;
  if (filters.sort_dir) params.sort_dir = filters.sort_dir;
  const res = await client.get<PaginatedResult<Artist>>("/artists", { params });
  return res.data;
}
