"use client";

import Link from "next/link";

import { useMe } from "@/features/auth/auth.hooks";
import { LogoutButton } from "@/features/auth/logout-button";
import { cn } from "@/shared/lib/cn";
import { Button } from "@/shared/ui/button";
import { NotificationBell } from "@/widgets/notification-bell/notification-bell";

export function Logo({ size = "lg" }: { size?: "lg" | "sm" }) {
  return (
    <Link href="/" className="flex items-center gap-2">
      <span
        className={cn(
          "flex rounded-full bg-gray-900 items-center justify-center",
          size === "lg" ? "h-6 w-6" : "h-5 w-5"
        )}
      >
        <span className={cn("rounded-full bg-white", size === "lg" ? "h-2 w-2" : "h-1.5 w-1.5")} />
      </span>
      <span className={cn("font-bold tracking-tight", size === "lg" ? "text-xl" : "text-lg")}>
        ReserveFlow
      </span>
    </Link>
  );
}

export function Header({ variant = "auto" }: { variant?: "public" | "auth" | "auto" }) {
  const me = useMe();
  const isAuthenticated = variant === "auth" || (variant === "auto" && Boolean(me.data));

  if (!isAuthenticated) {
    return (
      <header className="mx-auto flex w-full max-w-7xl items-center justify-between px-6 py-5">
        <Logo />
        <nav className="hidden gap-8 text-sm font-medium text-gray-600 md:flex">
          <Link className="transition-colors hover:text-gray-900" href="/events">
            Мероприятия
          </Link>
          <Link className="transition-colors hover:text-gray-900" href="/login">
            Войти
          </Link>
        </nav>
        <Link
          className="rounded-full bg-gray-900 px-6 py-3 text-sm font-medium text-white shadow-md transition-all hover:bg-gray-800 hover:shadow-lg"
          href="/register"
        >
          Регистрация
        </Link>
      </header>
    );
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b border-gray-100 bg-white px-6 py-4">
      <div className="mx-auto flex max-w-7xl items-center justify-between">
        <div className="flex items-center gap-8">
          <Logo size="sm" />
          <nav className="hidden gap-6 text-sm font-medium text-gray-500 md:flex">
            <Link className="transition-colors hover:text-gray-900" href="/events">
              Мероприятия
            </Link>
            <Link className="transition-colors hover:text-gray-900" href="/bookings">
              Мои брони
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-3">
          <NotificationBell />
          <div className="flex h-8 w-8 items-center justify-center rounded-full border border-gray-300 bg-gray-200 text-xs font-bold text-gray-700">
            {me.data?.name?.slice(0, 1).toUpperCase() ?? "U"}
          </div>
          <LogoutButton />
        </div>
      </div>
    </header>
  );
}
