import Link from "next/link";

import type { Event } from "@/entities/event/types";
import { cn } from "@/shared/lib/cn";
import { routes } from "@/shared/lib/routes";

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
function catLabel(category?: string | null) {
  if (!category) return "Мероприятие";
  return CATEGORY_LABELS[category] ?? category.replace(/[_-]+/g, " ");
}
function syntheticRating(id: string) {
  const code = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return (7.5 + (code % 18) / 10).toFixed(1);
}
function formatPrice(priceMin?: number | null) {
  if (!priceMin) return null;
  return `от ${priceMin.toLocaleString("ru-RU")} ₽`;
}

export function EventCard({ event }: { event: Event }) {
  const bg = coverColor(event.category);
  const symbol = coverSymbol(event.category);
  const rating = syntheticRating(event.id);
  const price = formatPrice(event.priceMin);

  const dateStr = event.startsAt
    ? new Intl.DateTimeFormat("ru-RU", {
        day: "numeric",
        month: "long",
        hour: "2-digit",
        minute: "2-digit"
      }).format(new Date(event.startsAt))
    : "Дата уточняется";

  const venue = event.venue?.name ?? event.venue?.address ?? "";

  return (
    <Link href={routes.event(event.id, event.title)} className="group block">
      <article className="overflow-hidden rounded-2xl border border-border bg-white shadow-card transition-shadow hover:shadow-card-hover">
        {/* Cover */}
        <div className={cn("relative flex h-[160px] items-center justify-center overflow-hidden", bg)}>
          {/* Real poster image — shown when available (KudaGo, TimePad, Yandex Afisha) */}
          {event.posterUrl && (
            <img
              src={event.posterUrl}
              alt=""
              aria-hidden
              referrerPolicy="no-referrer"
              className="absolute inset-0 h-full w-full object-cover opacity-60 transition-opacity group-hover:opacity-75"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = "none";
              }}
            />
          )}
          {/* Category badge */}
          <span className="absolute left-3 top-3 z-10 rounded-full bg-black/30 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-white/90 backdrop-blur-sm">
            {catLabel(event.category)}
          </span>
          {/* Age restriction badge */}
          {event.ageRestriction && (
            <span className="absolute bottom-3 left-3 z-10 rounded-full bg-black/40 px-2 py-0.5 text-[10px] font-bold text-white/80 backdrop-blur-sm">
              {event.ageRestriction}
            </span>
          )}
          {/* Rating badge */}
          <span className="absolute right-3 top-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-ok text-[11px] font-black text-white shadow">
            {rating}
          </span>
          {/* Symbol — shown when there's no photo */}
          {!event.posterUrl && (
            <span className="select-none text-4xl font-black text-white/25">{symbol}</span>
          )}
        </div>

        {/* Content */}
        <div className="p-4">
          <h3 className="line-clamp-2 text-[15px] font-bold leading-snug text-ink">
            {event.title}
          </h3>
          <p className="mt-1.5 text-[12px] text-mute">{dateStr}</p>
          {venue && (
            <p className="mt-0.5 line-clamp-1 text-[12px] text-mute-2">{venue}</p>
          )}
          <div className="mt-4 flex items-center justify-between">
            <span className="text-[12px] font-semibold text-ok">
              {price ?? (
                <span className="text-[11px] font-normal text-mute-2">
                  {event.status === "published" || event.status === "active" ? "Доступно" : "Скоро"}
                </span>
              )}
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
