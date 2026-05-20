"use client";

import { CalendarDays } from "lucide-react";
import Link from "next/link";
import { useDeferredValue, useEffect, useMemo, useState } from "react";

import type { Event } from "@/entities/event/types";
import { useEvents } from "@/features/event-list/event-list.hooks";
import { cn } from "@/shared/lib/cn";
import { routes } from "@/shared/lib/routes";

const EVENT_LIMIT = 60;

const CATEGORY_TABS = [
  { value: "all", label: "Все" },
  { value: "concert", label: "Концерты" },
  { value: "theatre", label: "Театр" },
  { value: "standup", label: "Стендап" },
  { value: "cinema", label: "Кино" },
  { value: "excursions", label: "Экскурсии" },
  { value: "quest", label: "Квест" },
  { value: "art", label: "Арт" },
  { value: "lectures", label: "Лекции" },
  { value: "masterclass", label: "Мастер-класс" },
  { value: "kids", label: "Детям" },
  { value: "party", label: "Вечеринки" },
  { value: "festival", label: "Фестивали" },
  { value: "show", label: "Шоу" },
  { value: "sport", label: "Спорт" },
  { value: "circus", label: "Цирк" }
];

const COVER_COLORS: Record<string, string> = {
  concert: "bg-[#1A1A2E]",
  festival: "bg-[#1B3A6B]",
  theatre: "bg-[#5C2E1A]",
  theater: "bg-[#5C2E1A]",
  theater_show: "bg-[#5C2E1A]",
  standup: "bg-[#1A3D2B]",
  cinema: "bg-[#1F1F35]",
  lecture: "bg-[#2D1B4E]",
  lectures: "bg-[#2D1B4E]",
  sport: "bg-[#1A2E1A]",
  kids: "bg-[#2E1A3D]",
  circus: "bg-[#3D1A1A]",
  art: "bg-[#2B1B4A]",
  excursions: "bg-[#1A3040]",
  masterclass: "bg-[#1F2B1A]",
  party: "bg-[#3D1A3D]",
  show: "bg-[#2A1A3D]",
  quest: "bg-[#2A2A1A]",
  default: "bg-[#1A1A2E]"
};

const COVER_SYMBOLS: Record<string, string> = {
  concert: "♪",
  festival: "○",
  theatre: "ЧЗ",
  theater: "ЧЗ",
  theater_show: "ЧЗ",
  standup: "☺",
  cinema: "III",
  lecture: "Ψ",
  lectures: "Ψ",
  sport: "◈",
  kids: "★",
  circus: "◉",
  art: "◇",
  excursions: "⊛",
  masterclass: "✦",
  party: "♦",
  show: "✧",
  quest: "◎",
  default: "◆"
};

const CATEGORY_LABELS: Record<string, string> = {
  concert: "Концерт",
  festival: "Фестиваль",
  theatre: "Театр",
  theater: "Театр",
  theater_show: "Театр",
  standup: "Стенд-ап",
  cinema: "Кино",
  lecture: "Лекция",
  lectures: "Лекция",
  sport: "Спорт",
  kids: "Детям",
  circus: "Цирк",
  art: "Арт",
  excursions: "Экскурсии",
  masterclass: "Мастер-класс",
  party: "Вечеринки",
  show: "Шоу",
  quest: "Квест"
};

function coverColor(category?: string | null) {
  return COVER_COLORS[category ?? ""] ?? COVER_COLORS.default;
}

function coverSymbol(category?: string | null) {
  return COVER_SYMBOLS[category ?? ""] ?? COVER_SYMBOLS.default;
}

function categoryLabel(category?: string | null) {
  if (!category) return "Мероприятие";
  return CATEGORY_LABELS[category] ?? category.replace(/[_-]+/g, " ");
}

function syntheticRating(id: string) {
  const code = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return (7.5 + (code % 18) / 10).toFixed(1);
}

// ─── Date strip helpers ─────────────────────────────────────────────────────

function getDayKey(value?: string | null) {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  return new Intl.DateTimeFormat("en-CA", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    timeZone: "Europe/Moscow"
  })
    .format(date)
    .replace(/\//g, "-");
}

function buildDayFilter(key: string, count: number) {
  const [y, m, d] = key.split("-").map(Number);
  const date = new Date(Date.UTC(y, m - 1, d, 9));
  const fmt = (opts: Intl.DateTimeFormatOptions) =>
    new Intl.DateTimeFormat("ru-RU", { timeZone: "Europe/Moscow", ...opts }).format(date);
  return {
    key,
    weekday: fmt({ weekday: "short" }).replace(".", ""),
    dayNumber: fmt({ day: "numeric" }),
    monthLabel: fmt({ month: "long" }),
    count
  };
}

function formatCardDate(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

function pluralEvents(n: number) {
  const r10 = n % 10;
  const r100 = n % 100;
  if (r10 === 1 && r100 !== 11) return "событие";
  if (r10 >= 2 && r10 <= 4 && (r100 < 12 || r100 > 14)) return "события";
  return "событий";
}

// ─── Main component ──────────────────────────────────────────────────────────

function todayKey() {
  return getDayKey(new Date().toISOString()) ?? "all";
}

export function HomeEventsShowcase() {
  const [activeCategory, setActiveCategory] = useState("all");
  const [activeDay, setActiveDay] = useState(() => todayKey());
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);

  // Only show events that are today or in the future.
  // The backend filter `onlyActual: true` adds: COALESCE(ends_at, starts_at) >= NOW()
  // This means the date strip auto-rolls forward each day — no manual refresh needed.
  const events = useEvents({ onlyActual: true, limit: EVENT_LIMIT, offset: 0 });

  const sorted = useMemo(
    () =>
      [...(events.data?.items ?? [])].sort((a, b) => {
        const ta = a.startsAt ? new Date(a.startsAt).getTime() : Number.MAX_SAFE_INTEGER;
        const tb = b.startsAt ? new Date(b.startsAt).getTime() : Number.MAX_SAFE_INTEGER;
        return ta !== tb ? ta - tb : a.title.localeCompare(b.title, "ru-RU");
      }),
    [events.data]
  );

  const dayOptions = useMemo(() => {
    const counts = new Map<string, number>();
    for (const ev of sorted) {
      const k = getDayKey(ev.startsAt);
      if (k) counts.set(k, (counts.get(k) ?? 0) + 1);
    }
    return [...counts.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .slice(0, 10)
      .map(([k, c]) => buildDayFilter(k, c));
  }, [sorted]);

  // If the initially-selected day (today) has no events yet, jump to the first
  // available day so the calendar is never empty.
  useEffect(() => {
    if (dayOptions.length > 0 && !dayOptions.find((d) => d.key === activeDay)) {
      setActiveDay(dayOptions[0].key);
    }
  }, [dayOptions, activeDay]);

  const filtered = useMemo(() => {
    const q = deferredQuery.trim().toLocaleLowerCase("ru-RU");
    return sorted.filter((ev) => {
      const matchQ =
        !q ||
        [ev.title, ev.description, ev.venue?.name, ev.venue?.address]
          .join(" ")
          .toLocaleLowerCase("ru-RU")
          .includes(q);
      const matchCat =
        activeCategory === "all" ||
        ev.category === activeCategory ||
        (activeCategory === "theatre" && (ev.category === "theater" || ev.category === "theater_show")) ||
        (activeCategory === "lectures" && ev.category === "lecture") ||
        (activeCategory === "lecture" && ev.category === "lectures");
      const matchDay = activeDay === "all" || getDayKey(ev.startsAt) === activeDay;
      return matchQ && matchCat && matchDay;
    });
  }, [sorted, deferredQuery, activeCategory, activeDay]);

  return (
    <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 pb-16 pt-8">
      {/* ── Заголовок афиши ── */}
      <div className="mb-6 flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
            АФИША · МОСКВА
          </p>
          <h1 className="mt-1 text-[28px] font-extrabold tracking-tight text-ink">
            События, на которые ещё можно успеть.{" "}
            <Link href={routes.eventsMap} className="text-mute underline underline-offset-2 hover:text-ink">
              Полный календарь
            </Link>
          </h1>
        </div>
        {/* Filter bar stub */}
        <div className="hidden items-center gap-2 rounded-full border border-border bg-white px-4 py-2 text-xs font-medium text-mute shadow-sm lg:flex">
          <span>Все фильтры</span>
          <span className="text-border">|</span>
          <span>Тип события</span>
          <span className="text-border">|</span>
          <span>Цена до 2 000 ₽</span>
          <span className="text-border">|</span>
          <span>Площадка</span>
          <span className="text-border">|</span>
          <span>Возраст</span>
        </div>
      </div>

      {/* ── Категории ── */}
      <div className="mb-6 flex gap-2 overflow-x-auto pb-1">
        {CATEGORY_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => setActiveCategory(tab.value)}
            className={cn(
              "shrink-0 rounded-full px-5 py-2 text-[13px] font-medium transition-all",
              activeCategory === tab.value
                ? "bg-ink text-white shadow"
                : "border border-border bg-white text-ink-2 hover:border-ink/30 hover:text-ink"
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* ── Сейчас обсуждают ── */}
      <div className="mb-2">
        <p className="text-[13px] font-bold text-ink">Сейчас обсуждают</p>
        <p className="text-xs text-mute">
          События с быстрым ростом броней за последние 24 часа
        </p>
      </div>

      {/* ── Дата-стрип ── */}
      {dayOptions.length > 0 && (
        <div className="mb-6 flex gap-2 overflow-x-auto pb-1">
          <button
            type="button"
            onClick={() => setActiveDay("all")}
            className={cn(
              "flex min-w-[72px] shrink-0 flex-col items-center rounded-xl border px-3 py-2 text-center transition-all",
              activeDay === "all"
                ? "border-ink bg-ink text-white"
                : "border-border bg-white text-ink-2 hover:border-ink/30"
            )}
          >
            <span className="text-[10px] uppercase tracking-wide opacity-70">Все</span>
            <span className="text-base font-black">∞</span>
            <span className="text-[10px]">даты</span>
          </button>
          {dayOptions.map((day) => (
            <button
              key={day.key}
              type="button"
              onClick={() => setActiveDay(day.key)}
              className={cn(
                "flex min-w-[72px] shrink-0 flex-col items-center rounded-xl border px-3 py-2 text-center transition-all",
                activeDay === day.key
                  ? "border-ink bg-ink text-white"
                  : "border-border bg-white text-ink-2 hover:border-ink/30"
              )}
            >
              <span className={cn("text-[10px] uppercase tracking-wide", activeDay === day.key ? "text-white/70" : "text-mute")}>
                {day.weekday}
              </span>
              <span className="text-xl font-black">{day.dayNumber}</span>
              <span className={cn("text-[10px]", activeDay === day.key ? "text-white/70" : "text-mute")}>
                {day.monthLabel}
              </span>
              <span className={cn("mt-0.5 text-[10px]", activeDay === day.key ? "text-white/60" : "text-mute-2")}>
                {day.count} {pluralEvents(day.count)}
              </span>
            </button>
          ))}
        </div>
      )}

      {/* ── Поиск ── */}
      <div className="mb-6">
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Поиск: концерт, площадка, артист…"
          className="w-full max-w-sm rounded-full border border-border bg-white px-5 py-2.5 text-[13px] text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none"
        />
      </div>

      {/* ── Статус загрузки ── */}
      {events.isLoading && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }, (_, i) => (
            <div key={i} className="h-[360px] animate-pulse rounded-2xl bg-white ring-1 ring-border" />
          ))}
        </div>
      )}

      {events.error && (
        <div className="rounded-2xl border border-err/20 bg-err-soft px-6 py-8 text-center text-sm text-err">
          Не удалось загрузить события. Проверьте backend и обновите страницу.
        </div>
      )}

      {!events.isLoading && !events.error && filtered.length === 0 && (
        <div className="rounded-2xl border border-border bg-white px-6 py-12 text-center shadow-card">
          <p className="text-base font-semibold text-ink">Ничего не найдено</p>
          <p className="mt-1 text-sm text-mute">Сбросьте фильтры и попробуйте снова</p>
          <button
            type="button"
            onClick={() => {
              setQuery("");
              setActiveCategory("all");
              setActiveDay("all");
            }}
            className="mt-4 rounded-full bg-ink px-5 py-2.5 text-sm font-medium text-white hover:bg-ink-2"
          >
            Сбросить фильтры
          </button>
        </div>
      )}

      {/* ── Карточки событий ── */}
      {!events.isLoading && !events.error && filtered.length > 0 && (
        <>
          {/* Первая карточка — большая */}
          <div className="mb-5 grid grid-cols-1 gap-5 lg:grid-cols-3">
            <FeaturedEventCard event={filtered[0]} />
            <div className="grid grid-cols-1 gap-5 lg:col-span-2 lg:grid-cols-2">
              {filtered.slice(1, 5).map((ev) => (
                <EventCard key={ev.id} event={ev} />
              ))}
            </div>
          </div>

          {/* Остальные карточки */}
          {filtered.length > 5 && (
            <>
              <p className="mb-4 text-[15px] font-bold text-ink">
                <CalendarDays className="mr-1.5 inline h-4 w-4 text-mute" />
                Все события
              </p>
              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
                {filtered.slice(5).map((ev) => (
                  <EventCard key={ev.id} event={ev} />
                ))}
              </div>
            </>
          )}

          <p className="mt-6 text-center text-sm text-mute">
            Показано {Math.min(filtered.length, filtered.length)} из {events.data?.total ?? filtered.length}{" "}
            {pluralEvents(events.data?.total ?? filtered.length)}
          </p>
        </>
      )}
    </main>
  );
}

// ─── FeaturedEventCard (большая карточка) ───────────────────────────────────
function FeaturedEventCard({ event }: { event: Event }) {
  const bg = coverColor(event.category);
  const symbol = coverSymbol(event.category);
  const rating = syntheticRating(event.id);
  const venue = event.venue?.name ?? event.venue?.address ?? "Площадка уточняется";
  const dateStr = event.startsAt ? formatCardDate(event.startsAt) : "Дата уточняется";

  return (
    <Link href={routes.event(event.id, event.title)} className="group block">
      <article className="overflow-hidden rounded-2xl border border-border bg-white shadow-card transition-shadow hover:shadow-card-hover">
        <div className={cn("relative flex h-[280px] flex-col items-center justify-center overflow-hidden", bg)}>
          {/* Poster image */}
          {event.posterUrl && (
            <img
              src={event.posterUrl}
              alt=""
              aria-hidden
              referrerPolicy="no-referrer"
              className="absolute inset-0 h-full w-full object-cover opacity-55 transition-opacity group-hover:opacity-65"
              onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
            />
          )}
          {/* Category badge */}
          <span className="absolute left-3 top-3 z-10 rounded-full bg-black/30 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-white/90 backdrop-blur-sm">
            {categoryLabel(event.category)}
            {event.ageRestriction ? ` · ${event.ageRestriction}` : ""}
          </span>
          {/* Rating */}
          <span className="absolute right-3 top-3 z-10 flex h-9 w-9 items-center justify-center rounded-full bg-ok text-[12px] font-black text-white shadow">
            {rating}
          </span>
          {/* Symbol — only when no poster */}
          {!event.posterUrl && (
            <span className="select-none text-6xl font-black text-white/25">{symbol}</span>
          )}
          {/* Bottom gradient overlay */}
          <div className="absolute inset-x-0 bottom-0 z-10 bg-gradient-to-t from-black/70 to-transparent p-4">
            <p className="text-xs text-white/70">{dateStr}</p>
            <h3 className="mt-1 line-clamp-2 text-xl font-black leading-snug text-white">
              {event.title}
            </h3>
            <p className="mt-1 line-clamp-1 text-xs text-white/70">{venue}</p>
          </div>
        </div>
        <div className="flex items-center justify-between p-4">
          <span className="text-xs text-mute">
            {event.venue?.city ?? "Москва"}
          </span>
          <span className="rounded-full bg-ink px-4 py-1.5 text-[12px] font-semibold text-white group-hover:bg-ink-2">
            Выбрать места →
          </span>
        </div>
      </article>
    </Link>
  );
}

// ─── EventCard (обычная карточка) ────────────────────────────────────────────
function EventCard({ event }: { event: Event }) {
  const bg = coverColor(event.category);
  const symbol = coverSymbol(event.category);
  const rating = syntheticRating(event.id);
  const dateStr = event.startsAt ? formatCardDate(event.startsAt) : "Дата уточняется";
  const venue = event.venue?.name ?? event.venue?.address ?? "Площадка уточняется";
  const price = event.priceMin
    ? `от ${event.priceMin.toLocaleString("ru-RU")} ₽`
    : null;

  return (
    <Link href={routes.event(event.id, event.title)} className="group block">
      <article className="overflow-hidden rounded-2xl border border-border bg-white shadow-card transition-shadow hover:shadow-card-hover">
        <div className={cn("relative flex h-[160px] items-center justify-center overflow-hidden", bg)}>
          {event.posterUrl && (
            <img
              src={event.posterUrl}
              alt=""
              aria-hidden
              referrerPolicy="no-referrer"
              className="absolute inset-0 h-full w-full object-cover opacity-60 transition-opacity group-hover:opacity-75"
              onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
            />
          )}
          <span className="absolute left-3 top-3 z-10 rounded-full bg-black/30 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-white/80 backdrop-blur-sm">
            {categoryLabel(event.category)}
          </span>
          <span className="absolute right-3 top-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-ok text-[11px] font-black text-white shadow">
            {rating}
          </span>
          {!event.posterUrl && (
            <span className="select-none text-4xl font-black text-white/25">{symbol}</span>
          )}
        </div>
        <div className="p-4">
          <h3 className="line-clamp-2 text-[14px] font-bold leading-snug text-ink">
            {event.title}
          </h3>
          <p className="mt-1 text-[12px] text-mute">{dateStr}</p>
          <p className="mt-0.5 line-clamp-1 text-[12px] text-mute-2">{venue}</p>
          <div className="mt-3 flex items-center justify-between">
            <span className="text-[12px] font-semibold text-ok">
              {price ?? <span className="text-[11px] font-normal text-mute-2">Доступно</span>}
            </span>
            <span className="text-[12px] font-semibold text-ink group-hover:underline">
              Подробнее →
            </span>
          </div>
        </div>
      </article>
    </Link>
  );
}
