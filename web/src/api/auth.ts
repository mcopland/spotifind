import type { User } from "../types";
import client from "./client";

export async function getMe(): Promise<User> {
  const res = await client.get<User>("/auth/me");
  return res.data;
}

export async function logout(): Promise<void> {
  await client.post("/auth/logout");
}
