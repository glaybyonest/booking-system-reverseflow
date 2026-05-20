"use client";

import { useMutation, useQuery } from "@tanstack/react-query";

import type { SeatLayout } from "@/entities/seat/types";
import { useMe } from "@/features/auth/auth.hooks";
import { LayoutEditor } from "@/features/admin-layout/layout-editor";
import { getHallLayoutState, saveHallLayout } from "@/shared/api/admin-layouts.api";
import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";

export function HallLayoutPage({ hallId }: { hallId: string }) {
  const me = useMe();
  const state = useQuery({
    queryKey: ["admin", "hall-layout", hallId],
    queryFn: () => getHallLayoutState(hallId),
    enabled: Boolean(hallId)
  });
  const save = useMutation({
    mutationFn: (layout: SeatLayout) => saveHallLayout(hallId, layout),
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
      title={`Fallback схема зала: ${state.data.name}`}
      description="Эта схема применяется ко всем сессиям зала без session override. После сохранения ReserveFlow пересчитает inventory и bookable-видимость связанных KudaGo-событий."
      initialLayout={state.data.layout}
      effectiveLayout={state.data.layout}
      layoutSource="hall"
      savePending={save.isPending}
      saveError={save.error}
      onSave={(layout) => save.mutate(layout)}
    />
  );
}
