import type {
  HallLayoutState,
  LayoutMutationResult,
  SessionLayoutState
} from "@/entities/layout/types";
import type { SeatLayout } from "@/entities/seat/types";
import { clientFetch } from "@/shared/api/client";

export async function getSessionLayoutState(sessionId: string) {
  return clientFetch<SessionLayoutState>(`admin/sessions/${sessionId}/layout`);
}

export async function saveSessionLayout(sessionId: string, layout: SeatLayout) {
  return clientFetch<LayoutMutationResult>(`admin/sessions/${sessionId}/layout`, {
    method: "PUT",
    json: { layout }
  });
}

export async function deleteSessionLayout(sessionId: string) {
  return clientFetch<LayoutMutationResult>(`admin/sessions/${sessionId}/layout`, {
    method: "DELETE"
  });
}

export async function getHallLayoutState(hallId: string) {
  return clientFetch<HallLayoutState>(`admin/halls/${hallId}/layout`);
}

export async function saveHallLayout(hallId: string, layout: SeatLayout) {
  return clientFetch<LayoutMutationResult>(`admin/halls/${hallId}/layout`, {
    method: "PUT",
    json: { layout }
  });
}
