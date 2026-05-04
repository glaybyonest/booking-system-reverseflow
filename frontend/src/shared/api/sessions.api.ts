import type { SeatMapResponse } from "@/entities/seat/types";
import type { Session } from "@/entities/session/types";
import { clientFetch } from "@/shared/api/client";
import { normalizeSeatMap, normalizeSession } from "@/shared/api/mappers";

export async function getSession(sessionId: string): Promise<Session> {
  return normalizeSession(await clientFetch(`sessions/${sessionId}`));
}

export async function getSessionSeats(sessionId: string): Promise<SeatMapResponse> {
  return normalizeSeatMap(await clientFetch(`sessions/${sessionId}/seats`));
}
