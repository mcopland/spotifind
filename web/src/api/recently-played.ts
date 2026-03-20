import type { PaginatedResult, RecentlyPlayedTrack } from "../types";
import client from "./client";

export async function getRecentlyPlayed(params: {
  page?: number;
  page_size?: number;
}): Promise<PaginatedResult<RecentlyPlayedTrack>> {
  const p: Record<string, string> = {};
  if (params.page) p.page = String(params.page);
  if (params.page_size) p.page_size = String(params.page_size);
  const res = await client.get<PaginatedResult<RecentlyPlayedTrack>>("/recently-played", { params: p });
  return res.data;
}
