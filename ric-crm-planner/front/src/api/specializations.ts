import client from "./client";
import { SPECIALIZATION_OPTIONS } from "../constants/specializations";
import type { Specialization } from "../types/event";

type BackendSpecialization = {
  id?: number | string;
  title?: string;
  name?: string;
};

let specializationsCache: Specialization[] | null = null;

function normalizeSpecialization(item: unknown): Specialization | null {
  if (!item || typeof item !== "object") return null;

  const raw = item as BackendSpecialization;
  const id = Number(raw.id);
  const title = String(raw.title ?? raw.name ?? "").trim();

  if (!Number.isFinite(id) || !title) return null;
  return { id, title };
}

export async function getSpecializations(): Promise<Specialization[]> {
  if (client.USE_MOCK) return SPECIALIZATION_OPTIONS;
  if (specializationsCache) return specializationsCache;

  try {
    const raw = await client.get<unknown[]>("/api/users/specializations/");
    const loaded = Array.isArray(raw)
      ? raw.map(normalizeSpecialization).filter((item): item is Specialization => Boolean(item))
      : [];

    specializationsCache = loaded.length ? loaded : SPECIALIZATION_OPTIONS;
    return specializationsCache;
  } catch {
    return SPECIALIZATION_OPTIONS;
  }
}
