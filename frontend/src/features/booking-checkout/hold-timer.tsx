"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import { minutesSeconds } from "@/shared/lib/date";
import { cn } from "@/shared/lib/cn";

export function HoldTimer({ expiresAt, onExpired }: { expiresAt?: string | null; onExpired?: () => void }) {
  const [now, setNow] = useState(() => Date.now());
  const fired = useRef(false);
  const remaining = useMemo(() => {
    if (!expiresAt) return 0;
    return Math.max(0, Math.floor((new Date(expiresAt).getTime() - now) / 1000));
  }, [expiresAt, now]);

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, []);

  useEffect(() => {
    if (remaining <= 0 && expiresAt && !fired.current) {
      fired.current = true;
      onExpired?.();
    }
  }, [expiresAt, onExpired, remaining]);

  const isUrgent = remaining < 60;
  const isWarning = remaining < 180;

  return (
    <div
      className={cn(
        "rounded-2xl border p-4 text-center transition-colors",
        isUrgent
          ? "border-err/30 bg-err-soft"
          : isWarning
            ? "border-warn/30 bg-warn-soft"
            : "border-border bg-bg"
      )}
    >
      <p
        className={cn(
          "text-[10px] font-bold uppercase tracking-widest",
          isUrgent ? "text-err-fg" : isWarning ? "text-warn-fg" : "text-mute-2"
        )}
      >
        Время на оплату
      </p>
      <div
        className={cn(
          "mt-1 font-mono text-3xl font-extrabold tracking-tight",
          isUrgent ? "text-err" : isWarning ? "text-warn" : "text-ink"
        )}
      >
        {minutesSeconds(remaining)}
      </div>
    </div>
  );
}
