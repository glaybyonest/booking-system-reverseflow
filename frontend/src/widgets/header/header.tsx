"use client";

import { LogOut, MapPin, Search } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { useLogout, useMe } from "@/features/auth/auth.hooks";
import { cn } from "@/shared/lib/cn";
import { NotificationBell } from "@/widgets/notification-bell/notification-bell";

export function Logo({ size = "lg", light = false }: { size?: "lg" | "sm"; light?: boolean }) {
  return (
    <Link href="/" className="flex shrink-0 items-center gap-2">
      <span
        className={cn(
          "flex items-center justify-center rounded-full",
          light ? "bg-white" : "bg-ink",
          size === "lg" ? "h-6 w-6" : "h-5 w-5"
        )}
      >
        <span
          className={cn(
            "rounded-full",
            light ? "bg-ink" : "bg-white",
            size === "lg" ? "h-2 w-2" : "h-1.5 w-1.5"
          )}
        />
      </span>
      <span
        className={cn(
          "font-bold tracking-tight",
          light ? "text-white" : "text-ink",
          size === "lg" ? "text-[15px]" : "text-sm"
        )}
      >
        ReserveFlow
      </span>
    </Link>
  );
}

function UserAvatar({ name }: { name?: string }) {
  const initials = name
    ? name
        .split(" ")
        .map((p) => p[0])
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : "U";
  return (
    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-ink text-[11px] font-bold text-white">
      {initials}
    </div>
  );
}

export function Header({ variant = "auto" }: { variant?: "public" | "auth" | "auto" }) {
  const me = useMe();
  const logout = useLogout();
  const router = useRouter();
  const isAuthenticated = variant === "auth" || (variant === "auto" && Boolean(me.data));

  function handleLogout() {
    logout.mutate(undefined, {
      onSuccess: () => router.push("/login")
    });
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-white">
      <div className="mx-auto flex h-14 max-w-[1440px] items-center gap-4 px-6">
        <Logo size="sm" />

        <div className="flex flex-1 items-center justify-center">
          <div className="relative w-full max-w-[340px]">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-mute-2" />
            <input
              type="search"
              placeholder="События, артисты, площадки"
              className="w-full rounded-lg border border-border bg-bg py-2 pl-9 pr-4 text-sm text-ink outline-none placeholder:text-mute-2 focus:border-ink/30 focus:ring-2 focus:ring-ink/8"
            />
          </div>
        </div>

        <nav className="flex items-center gap-1">
          <Link
            href="/events"
            className="rounded-lg px-3 py-2 text-sm font-medium text-ink transition-colors hover:bg-bg"
          >
            Мероприятия
          </Link>
          {isAuthenticated ? (
            <Link
              href="/bookings"
              className="rounded-lg px-3 py-2 text-sm font-medium text-mute transition-colors hover:bg-bg hover:text-ink"
            >
              Мои брони
            </Link>
          ) : (
            <Link
              href="/login"
              className="rounded-lg px-3 py-2 text-sm font-medium text-mute transition-colors hover:bg-bg hover:text-ink"
            >
              Войти
            </Link>
          )}
        </nav>

        <div className="flex items-center gap-2 border-l border-border pl-4">
          <button className="flex items-center gap-1 rounded-lg px-2 py-2 text-xs font-medium text-mute transition-colors hover:bg-bg hover:text-ink">
            <MapPin className="h-3.5 w-3.5" />
            <span>Москва</span>
          </button>

          {isAuthenticated ? (
            <>
              <NotificationBell />
              <UserAvatar name={me.data?.name} />
              <button
                onClick={handleLogout}
                disabled={logout.isPending}
                title="Выйти"
                className="flex h-8 w-8 items-center justify-center rounded-full text-mute transition-colors hover:bg-bg hover:text-ink disabled:opacity-50"
              >
                <LogOut className="h-4 w-4" />
              </button>
            </>
          ) : (
            <Link
              href="/register"
              className="rounded-full bg-ink px-4 py-2 text-xs font-semibold text-white transition-colors hover:bg-ink-2"
            >
              Регистрация
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
