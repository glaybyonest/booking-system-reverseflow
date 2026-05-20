"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";

import { friendlyApiError } from "@/shared/api/errors";
import { toSafeRedirectPath } from "@/shared/lib/url";
import { Alert } from "@/shared/ui/alert";
import { LoginInput, loginSchema } from "@/features/auth/auth.schemas";
import { useLogin } from "@/features/auth/auth.hooks";

export function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useLogin();
  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "demo@example.com", password: "Password123!" }
  });

  const onSubmit = form.handleSubmit((values) => {
    login.mutate(values, {
      onSuccess: () => router.push(toSafeRedirectPath(searchParams.get("redirect")))
    });
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      {login.error ? <Alert variant="error">{friendlyApiError(login.error)}</Alert> : null}

      <div className="space-y-1.5">
        <label className="block text-[13px] font-medium text-ink">Email</label>
        <input
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          className="w-full rounded-xl border border-border bg-white px-4 py-3 text-sm text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
          {...form.register("email")}
        />
        {form.formState.errors.email && (
          <p className="text-xs text-err">{form.formState.errors.email.message}</p>
        )}
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <label className="block text-[13px] font-medium text-ink">Пароль</label>
          <span className="text-[12px] text-mute hover:text-ink cursor-pointer">Забыли пароль?</span>
        </div>
        <input
          type="password"
          autoComplete="current-password"
          placeholder="••••••••"
          className="w-full rounded-xl border border-border bg-white px-4 py-3 text-sm text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
          {...form.register("password")}
        />
        {form.formState.errors.password && (
          <p className="text-xs text-err">{form.formState.errors.password.message}</p>
        )}
      </div>

      <label className="flex items-center gap-2 text-[13px] text-mute">
        <input type="checkbox" className="rounded border-border accent-ink" />
        Запомнить меня
      </label>

      <button
        type="submit"
        disabled={login.isPending}
        className="w-full rounded-full bg-ink py-3 text-sm font-bold text-white transition-colors hover:bg-ink-2 disabled:opacity-50"
      >
        {login.isPending ? "Входим..." : "Войти"}
      </button>

      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border" />
        </div>
        <div className="relative flex justify-center text-xs text-mute">
          <span className="bg-bg px-3">или</span>
        </div>
      </div>

      <button
        type="button"
        className="w-full rounded-full border border-border bg-white py-3 text-sm font-medium text-ink hover:border-ink/30 hover:text-ink"
      >
        Продолжить через корпоративный SSO
      </button>

      <p className="text-center text-[13px] text-mute">
        Нет аккаунта?{" "}
        <Link href="/register" className="font-semibold text-ink hover:underline">
          Создать аккаунт
        </Link>
      </p>
    </form>
  );
}
