"use client";

import { useMutation, useQuery } from "@tanstack/react-query";

import type { SeatLayout } from "@/entities/seat/types";
import { useMe } from "@/features/auth/auth.hooks";
import { LayoutEditor } from "@/features/admin-layout/layout-editor";
import {
  deleteSessionLayout,
  getSessionLayoutState,
  saveSessionLayout
} from "@/shared/api/admin-layouts.api";
import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";

export function SessionLayoutPage({ sessionId }: { sessionId: string }) {
  const me = useMe();
  const state = useQuery({
    queryKey: ["admin", "session-layout", sessionId],
    queryFn: () => getSessionLayoutState(sessionId),
    enabled: Boolean(sessionId)
  });
  const save = useMutation({
    mutationFn: (layout: SeatLayout) => saveSessionLayout(sessionId, layout),
    onSuccess: () => state.refetch()
  });
  const remove = useMutation({
    mutationFn: () => deleteSessionLayout(sessionId),
    onSuccess: () => state.refetch()
  });

  if (me.isLoading || state.isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner />
      </div>
    );
  }
  if (me.error) return <Alert variant="error">{friendlyApiError(me.error)}</Alert>;
  if (me.data?.role !== "admin") {
    return <Alert variant="error">Доступ к редактору схем есть только у администраторов.</Alert>;
  }
  if (state.error) return <Alert variant="error">{friendlyApiError(state.error)}</Alert>;
  if (!state.data) return null;

  return (
    <LayoutEditor
      title={`Схема сеанса: ${state.data.eventTitle}`}
      description="Session override имеет приоритет над fallback зала и сразу влияет на публичную видимость события, если после материализации появляются bookable места."
      initialLayout={state.data.layout}
      fallbackLayout={state.data.fallbackLayout}
      effectiveLayout={state.data.effectiveLayout}
      layoutSource={state.data.layoutSource}
      hallId={state.data.hall?.id}
      savePending={save.isPending}
      deletePending={remove.isPending}
      saveError={save.error}
      deleteError={remove.error}
      onSave={(layout) => save.mutate(layout)}
      onDelete={state.data.layout ? () => remove.mutate() : undefined}
    />
  );
}
