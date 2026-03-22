import type { PaginatedResult, TimeRange, TopArtist, TopTrack } from "../types";
import client from "./client";

export async function getTopTracks(params: {
  time_range: TimeRange;
  page?: number;
  page_size?: number;
}): Promise<PaginatedResult<TopTrack>> {
  const p: Record<string, string> = { time_range: params.time_range };
  if (params.page) p.page = String(params.page);
  if (params.page_size) p.page_size = String(params.page_size);
  const res = await client.get<PaginatedResult<TopTrack>>("/top/tracks", { params: p });
  return res.data;
}

export async function getTopArtists(params: {
  time_range: TimeRange;
  page?: number;
  page_size?: number;
}): Promise<PaginatedResult<TopArtist>> {
  const p: Record<string, string> = { time_range: params.time_range };
  if (params.page) p.page = String(params.page);
  if (params.page_size) p.page_size = String(params.page_size);
  const res = await client.get<PaginatedResult<TopArtist>>("/top/artists", { params: p });
  return res.data;
}
