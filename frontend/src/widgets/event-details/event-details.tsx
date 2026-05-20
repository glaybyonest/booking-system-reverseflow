"use client";

import { CalendarDays, MapPin, Train } from "lucide-react";

import { useEvent } from "@/features/event-list/event-list.hooks";
import { SessionList } from "@/features/session-list/session-list";
import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime, formatTimeRange } from "@/shared/lib/date";
import { Alert } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";
import { EventMiniMap } from "@/widgets/event-map/event-mini-map";

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

function syntheticRating(id: string) {
  const code = id.split("").reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return (7.5 + (code % 18) / 10).toFixed(1);
}

function formatPrice(priceMin?: number | null, priceMax?: number | null) {
  if (!priceMin) return null;
  const minStr = priceMin.toLocaleString("ru-RU");
  if (priceMax && priceMax > priceMin) {
    const maxStr = priceMax.toLocaleString("ru-RU");
    return `${minStr} — ${maxStr} ₽`;
  }
  return `от ${minStr} ₽`;
}

export function EventDetails({ eventId }: { eventId: string }) {
  const event = useEvent(eventId);

  if (event.isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner />
      </div>
    );
  }
  if (event.error) return <Alert variant="error">{friendlyApiError(event.error)}</Alert>;
  if (!event.data) return null;

  const bg = COVER_COLORS[event.data.category ?? ""] ?? COVER_COLORS.default;
  const symbol = COVER_SYMBOLS[event.data.category ?? ""] ?? COVER_SYMBOLS.default;
  const rating = syntheticRating(event.data.id);
  const catLabel = CATEGORY_LABELS[event.data.category ?? ""] ?? (event.data.category ?? "Мероприятие");
  const sessions = event.data.sessions ?? [];
  const priceStr = formatPrice(event.data.priceMin, event.data.priceMax);

  const dateStr = event.data.startsAt
    ? new Intl.DateTimeFormat("ru-RU", {
        day: "numeric",
        month: "long",
        hour: "2-digit",
        minute: "2-digit"
      }).format(new Date(event.data.startsAt))
    : "Дата уточняется";

  const timeRange = event.data.startsAt && event.data.endsAt
    ? formatTimeRange(event.data.startsAt, event.data.endsAt)
    : null;

  const nearestSession = sessions.find((s) => s.isBookable) ?? sessions[0];

  // Prefer long description, fall back to short
  const fullDescription = event.data.longDescription ?? event.data.description;

  return (
    <div className="grid gap-8 lg:grid-cols-12">
      {/* ── Left: cover + info ── */}
      <section className="space-y-6 lg:col-span-7">
        {/* Cover */}
        <div className={`relative flex h-[260px] items-center justify-center overflow-hidden rounded-2xl ${bg}`}>
          {/* Real poster image when available */}
          {event.data.posterUrl && (
            <img
              src={event.data.posterUrl}
              alt=""
              aria-hidden
              referrerPolicy="no-referrer"
              className="absolute inset-0 h-full w-full object-cover opacity-55"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = "none";
              }}
            />
          )}
          <span className="absolute left-4 top-4 z-10 rounded-full bg-black/30 px-3 py-1 text-[11px] font-bold uppercase tracking-wide text-white/90 backdrop-blur-sm">
            {catLabel}
          </span>
          {/* Age restriction */}
          {event.data.ageRestriction && (
            <span className="absolute bottom-4 left-4 z-10 rounded-full bg-black/40 px-3 py-1 text-[11px] font-bold text-white backdrop-blur-sm">
              {event.data.ageRestriction}
            </span>
          )}
          <span className="absolute right-4 top-4 z-10 flex h-9 w-9 items-center justify-center rounded-full bg-ok text-[12px] font-black text-white shadow">
            {rating}
          </span>
          {!event.data.posterUrl && (
            <span className="select-none text-7xl font-black text-white/20">{symbol}</span>
          )}
        </div>

        {/* Title & meta */}
        <div className="rounded-2xl border border-border bg-white p-6 shadow-card">
          <h1 className="text-2xl font-extrabold leading-snug text-ink md:text-3xl">
            {event.data.title}
          </h1>

          <div className="mt-4 space-y-2.5">
            <div className="flex items-start gap-2.5 text-sm text-mute">
              <CalendarDays className="mt-0.5 h-4 w-4 shrink-0 text-mute-2" />
              <span>
                {dateStr}
                {timeRange ? ` · ${timeRange}` : ""}
              </span>
            </div>
            {event.data.venue && (
              <div className="flex items-start gap-2.5 text-sm text-mute">
                <MapPin className="mt-0.5 h-4 w-4 shrink-0 text-mute-2" />
                <span>
                  {event.data.venue.name}
                  {event.data.venue.address ? `, ${event.data.venue.address}` : ""}
                </span>
              </div>
            )}
            {/* Metro stations */}
            {event.data.venue?.metroStations && event.data.venue.metroStations.length > 0 && (
              <div className="flex items-start gap-2.5 text-sm text-mute">
                <Train className="mt-0.5 h-4 w-4 shrink-0 text-mute-2" />
                <span>{event.data.venue.metroStations.join(", ")}</span>
              </div>
            )}
          </div>

          {/* Price */}
          {priceStr && (
            <div className="mt-4 inline-flex items-center rounded-full border border-ok/20 bg-ok/5 px-4 py-1.5">
              <span className="text-sm font-bold text-ok">{priceStr}</span>
            </div>
          )}

          {/* Description */}
          {fullDescription && (
            <div className="mt-5">
              <p className="text-sm leading-relaxed text-mute">{fullDescription}</p>
            </div>
          )}

          {/* Tags */}
          {event.data.tags && event.data.tags.length > 0 && (
            <div className="mt-5 flex flex-wrap gap-2">
              {event.data.tags.slice(0, 8).map((tag) => (
                <span
                  key={tag}
                  className="rounded-full border border-border bg-bg px-3 py-1 text-[11px] font-medium text-mute-2"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

        </div>

        {/* Mini map */}
        {event.data.venue?.latitude && event.data.venue?.longitude && (
          <div className="overflow-hidden rounded-2xl border border-border bg-white shadow-card">
            <div className="px-5 pt-5">
              <p className="text-[11px] font-bold uppercase tracking-widest text-mute-2">Локация</p>
              <p className="mt-0.5 text-sm font-semibold text-ink">{event.data.venue.name}</p>
              {event.data.venue.address && (
                <p className="mt-0.5 text-xs text-mute">{event.data.venue.address}</p>
              )}
              {event.data.venue.metroStations && event.data.venue.metroStations.length > 0 && (
                <p className="mt-1 flex items-center gap-1.5 text-xs text-mute-2">
                  <Train className="h-3 w-3" />
                  {event.data.venue.metroStations.join(" · ")}
                </p>
              )}
            </div>
            <div className="mt-4">
              <EventMiniMap
                latitude={event.data.venue.latitude}
                longitude={event.data.venue.longitude}
                title={event.data.title}
              />
            </div>
          </div>
        )}
      </section>

      {/* ── Right: quick-info + sessions ── */}
      <aside className="space-y-5 lg:col-span-5">
        {/* Quick-info sidebar */}
        <div className="rounded-2xl border border-border bg-white p-5 shadow-card">
          <p className="text-[11px] font-bold uppercase tracking-widest text-mute-2">О мероприятии</p>

          <div className="mt-4 grid grid-cols-2 gap-4">
            {nearestSession && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Ближайший сеанс</p>
                <p className="mt-1 text-sm font-semibold text-ink">
                  {nearestSession.startsAt
                    ? new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" }).format(new Date(nearestSession.startsAt))
                    : "Скоро"}
                </p>
              </div>
            )}

            {event.data.venue?.name && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Площадка</p>
                <p className="mt-1 text-sm font-semibold text-ink">{event.data.venue.name}</p>
              </div>
            )}

            {event.data.venue?.city && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Город</p>
                <p className="mt-1 text-sm font-semibold text-ink">{event.data.venue.city}</p>
              </div>
            )}

            <div>
              <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Статус</p>
              <p className="mt-1 text-sm font-semibold text-ink">
                {event.data.status === "published" || event.data.status === "active"
                  ? "Доступно"
                  : "Скоро"}
              </p>
            </div>

            {event.data.ageRestriction && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Возраст</p>
                <p className="mt-1 text-sm font-semibold text-ink">{event.data.ageRestriction}</p>
              </div>
            )}

            {priceStr && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Цена</p>
                <p className="mt-1 text-sm font-semibold text-ok">{priceStr}</p>
              </div>
            )}

            {event.data.venue?.venueTypeName && (
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">Тип площадки</p>
                <p className="mt-1 text-sm font-semibold text-ink">{event.data.venue.venueTypeName}</p>
              </div>
            )}
          </div>

        </div>

        {/* Sessions */}
        <div>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-extrabold text-ink">Сеансы</h2>
            {sessions.length > 0 && (
              <span className="text-[11px] font-medium text-mute-2">{sessions.length} шт.</span>
            )}
          </div>
          <SessionList sessions={sessions} />
        </div>
      </aside>
    </div>
  );
}
