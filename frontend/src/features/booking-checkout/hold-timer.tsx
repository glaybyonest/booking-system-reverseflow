"use client";

import { useEffect, useMemo, useRef, useState } from "react";

import { minutesSeconds } from "@/shared/lib/date";

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

  return (
    <div className="rounded-2xl border border-gray-200 bg-[#F8F9FA] p-5 text-center">
      <p className="text-xs font-bold uppercase tracking-wider text-gray-400">Время на оплату</p>
      <div className="mt-2 text-4xl font-extrabold tracking-tight">{minutesSeconds(remaining)}</div>
    </div>
  );
}
