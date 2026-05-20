import { ArrowRight } from "lucide-react";
import Link from "next/link";

import type { Session } from "@/entities/session/types";
import { formatDateTime, formatTimeRange } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";

export function SessionCard({ session, isAdmin = false }: { session: Session; isAdmin?: boolean }) {
  const isBookable = Boolean(session.isBookable);

  return (
    <div className="flex items-center justify-between rounded-2xl border border-border bg-white px-5 py-4 shadow-card transition-shadow hover:shadow-card-hover">
      <div className="min-w-0 flex-1">
        {/* Status badge */}
        <span
          className={`inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide ${
            session.status === "scheduled"
              ? "bg-ok-soft text-ok-fg"
              : "bg-bg text-mute-2"
          }`}
        >
          {session.status === "scheduled" ? "Запланировано" : session.status}
        </span>

        <p className="mt-2 text-[15px] font-bold text-ink">
          {formatDateTime(session.startsAt)}
        </p>
        <p className="mt-0.5 text-xs text-mute">
          {formatTimeRange(session.startsAt, session.endsAt)}
          {(session.hall?.name ?? session.hallName) ? ` · ${session.hall?.name ?? session.hallName}` : ""}
        </p>
      </div>

      <div className="ml-4 flex shrink-0 flex-col items-end gap-2">
        {isBookable ? (
          <Link
            href={routes.session(session.id)}
            className="flex items-center gap-1.5 rounded-full bg-ink px-4 py-2 text-[13px] font-bold text-white transition-colors hover:bg-ink-2"
          >
            Купить
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        ) : (
          <span className="rounded-full border border-border bg-bg px-4 py-2 text-[13px] font-medium text-mute">
            Скоро
          </span>
        )}

        {isAdmin && (
          <div className="flex flex-wrap justify-end gap-1.5">
            <Link
              href={routes.adminSessionLayout(session.id)}
              className="rounded-full border border-border bg-white px-3 py-1 text-[11px] font-medium text-ink hover:border-ink/30"
            >
              Макет сеанса
            </Link>
            {session.hall?.id && (
              <Link
                href={routes.adminHallLayout(session.hall.id)}
                className="rounded-full border border-border bg-white px-3 py-1 text-[11px] font-medium text-mute hover:border-ink/30 hover:text-ink"
              >
                Макет зала
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
