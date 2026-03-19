import type { Stats, SyncJob } from "../types";
import client from "./client";

export async function triggerSync(): Promise<{ job_id: number }> {
  const res = await client.post<{ job_id: number }>("/sync");
  return res.data;
}

export async function getSyncStatus(): Promise<SyncJob> {
  const res = await client.get<SyncJob>("/sync/status");
  return res.data;
}

export async function getGenres(): Promise<string[]> {
  const res = await client.get<string[]>("/genres");
  return res.data;
}

export async function getStats(): Promise<Stats> {
  const res = await client.get<Stats>("/stats");
  return res.data;
}
