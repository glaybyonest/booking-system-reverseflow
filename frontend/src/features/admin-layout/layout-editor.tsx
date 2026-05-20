"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import type { Seat, SeatLayout } from "@/entities/seat/types";
import { friendlyApiError } from "@/shared/api/errors";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { ExactSeatMap } from "@/widgets/seat-map/exact-seat-map";

type LayoutEditorProps = {
  title: string;
  description: string;
  initialLayout?: SeatLayout | null;
  fallbackLayout?: SeatLayout | null;
  effectiveLayout?: SeatLayout | null;
  layoutSource: string;
  hallId?: string | null;
  savePending?: boolean;
  deletePending?: boolean;
  saveError?: unknown;
  deleteError?: unknown;
  onSave(layout: SeatLayout): void;
  onDelete?: () => void;
};

export function LayoutEditor({
  title,
  description,
  initialLayout,
  fallbackLayout,
  effectiveLayout,
  layoutSource,
  hallId,
  savePending = false,
  deletePending = false,
  saveError,
  deleteError,
  onSave,
  onDelete
}: LayoutEditorProps) {
  const seedLayout = initialLayout ?? effectiveLayout ?? buildTemplateLayout(6, 12);
  const [rawLayout, setRawLayout] = useState(() => JSON.stringify(seedLayout, null, 2));
  const [templateRows, setTemplateRows] = useState("6");
  const [templateSeats, setTemplateSeats] = useState("12");

  useEffect(() => {
    setRawLayout(JSON.stringify(initialLayout ?? effectiveLayout ?? buildTemplateLayout(6, 12), null, 2));
  }, [effectiveLayout, initialLayout]);

  const parsedState = useMemo(() => {
    try {
      return { layout: JSON.parse(rawLayout) as SeatLayout, error: null as string | null };
    } catch (error) {
      return {
        layout: undefined,
        error: error instanceof Error ? error.message : "Не удалось разобрать JSON"
      };
    }
  }, [rawLayout]);
  const parsedLayout = parsedState.layout;
  const parseError = parsedState.error;

  const previewSeats = useMemo(() => buildPreviewSeats(parsedLayout ?? effectiveLayout ?? fallbackLayout), [
    effectiveLayout,
    fallbackLayout,
    parsedLayout
  ]);

  return (
    <div className="space-y-6">
      <Card className="space-y-4 p-6">
        <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div className="space-y-2">
            <h1 className="text-3xl font-bold">{title}</h1>
            <p className="max-w-3xl text-sm text-gray-500">{description}</p>
          </div>
          {hallId ? (
            <Link href={routes.adminHallLayout(hallId)}>
              <Button variant="secondary">Редактировать fallback зала</Button>
            </Link>
          ) : null}
        </div>

        <div className="flex flex-wrap gap-3 text-sm text-gray-500">
          <span className="rounded-full border border-gray-200 bg-gray-50 px-4 py-2">
            Источник effective layout: {layoutSource}
          </span>
          {initialLayout ? (
            <span className="rounded-full border border-gray-200 bg-gray-50 px-4 py-2">
              У сессии есть собственная схема
            </span>
          ) : null}
          {!initialLayout && fallbackLayout ? (
            <span className="rounded-full border border-gray-200 bg-gray-50 px-4 py-2">
              Сейчас используется fallback зала
            </span>
          ) : null}
        </div>

        {saveError ? <Alert variant="error">{friendlyApiError(saveError)}</Alert> : null}
        {deleteError ? <Alert variant="error">{friendlyApiError(deleteError)}</Alert> : null}
        {parseError ? <Alert variant="error">JSON схемы некорректен: {parseError}</Alert> : null}
      </Card>

      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card className="space-y-4 p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-xl font-bold">Редактор layout JSON</h2>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="secondary"
                onClick={() => {
                  const rows = clampTemplateValue(templateRows, 1, 26, 6);
                  const seatsPerRow = clampTemplateValue(templateSeats, 1, 40, 12);
                  setRawLayout(JSON.stringify(buildTemplateLayout(rows, seatsPerRow), null, 2));
                }}
              >
                Сгенерировать шаблон
              </Button>
              {onDelete ? (
                <Button variant="danger" disabled={deletePending} onClick={onDelete}>
                  Удалить override
                </Button>
              ) : null}
              <Button
                disabled={!parsedLayout || savePending}
                onClick={() => {
                  if (parsedLayout) {
                    onSave(parsedLayout);
                  }
                }}
              >
                Сохранить схему
              </Button>
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            <label className="space-y-1 text-sm text-gray-500">
              <span>Рядов в шаблоне</span>
              <Input value={templateRows} onChange={(event) => setTemplateRows(event.target.value)} />
            </label>
            <label className="space-y-1 text-sm text-gray-500">
              <span>Мест в ряду</span>
              <Input value={templateSeats} onChange={(event) => setTemplateSeats(event.target.value)} />
            </label>
          </div>

          <label className="block space-y-2 text-sm text-gray-500">
            <span>Layout JSON</span>
            <textarea
              className="min-h-[520px] w-full rounded-[1.5rem] border border-gray-200 bg-[#0B1120] p-4 font-mono text-sm text-slate-100 outline-none transition focus:border-gray-900 focus:ring-2 focus:ring-gray-900/10"
              spellCheck={false}
              value={rawLayout}
              onChange={(event) => setRawLayout(event.target.value)}
            />
          </label>
        </Card>

        <Card className="space-y-4 p-6">
          <div>
            <h2 className="text-xl font-bold">Live preview</h2>
            <p className="mt-1 text-sm text-gray-500">
              Точный вид зала формируется из этого JSON. Для one-to-one схемы правьте координаты и состав мест прямо здесь.
            </p>
          </div>
          {parsedLayout ?? effectiveLayout ?? fallbackLayout ? (
            <ExactSeatMap
              layout={parsedLayout ?? effectiveLayout ?? fallbackLayout}
              seats={previewSeats}
              interactive={false}
            />
          ) : (
            <Alert variant="info">Схема пока не задана. Сгенерируйте шаблон или вставьте готовый layout JSON.</Alert>
          )}
        </Card>
      </div>
    </div>
  );
}

function buildPreviewSeats(layout?: SeatLayout | null): Seat[] {
  if (!layout) {
    return [];
  }
  return layout.seats.map((seat) => ({
    seatId: seat.key,
    layoutKey: seat.key,
    row: seat.row,
    number: seat.number,
    label: seat.label ?? `${seat.row}-${seat.number}`,
    status: "available",
    price: seat.price
  }));
}

function clampTemplateValue(value: string, min: number, max: number, fallback: number) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return fallback;
  return Math.max(min, Math.min(max, Math.floor(parsed)));
}

function buildTemplateLayout(rows: number, seatsPerRow: number): SeatLayout {
  const layoutRows = Array.from({ length: rows }).map((_, rowIndex) =>
    String.fromCharCode("A".charCodeAt(0) + rowIndex)
  );

  const seats = layoutRows.flatMap((row, rowIndex) =>
    Array.from({ length: seatsPerRow }).map((_, seatIndex) => {
      const number = seatIndex + 1;
      return {
        key: `${row}-${number}`,
        label: `${row}-${number}`,
        row,
        number,
        x: 140 + seatIndex * 48,
        y: 170 + rowIndex * 56,
        price: rowIndex < 2 ? 3500 : rowIndex < 4 ? 2800 : 2200
      };
    })
  );

  return {
    version: 1,
    canvas: {
      width: Math.max(840, 220 + seatsPerRow * 48),
      height: Math.max(520, 250 + rows * 56)
    },
    stage: {
      label: "Сцена",
      x: 160,
      y: 44,
      width: Math.max(300, seatsPerRow * 34),
      height: 58
    },
    seats
  };
}
