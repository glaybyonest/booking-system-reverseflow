"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";

import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
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
      onSuccess: () => router.push(searchParams.get("redirect") ?? "/events")
    });
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      {login.error ? <Alert variant="error">{friendlyApiError(login.error)}</Alert> : null}
      <label className="block space-y-2 text-sm font-medium">
        <span>Email</span>
        <Input type="email" autoComplete="email" {...form.register("email")} />
        {form.formState.errors.email ? (
          <span className="text-xs text-red-600">{form.formState.errors.email.message}</span>
        ) : null}
      </label>
      <label className="block space-y-2 text-sm font-medium">
        <span>Пароль</span>
        <Input type="password" autoComplete="current-password" {...form.register("password")} />
        {form.formState.errors.password ? (
          <span className="text-xs text-red-600">{form.formState.errors.password.message}</span>
        ) : null}
      </label>
      <Button className="w-full" size="lg" type="submit" disabled={login.isPending}>
        {login.isPending ? "Входим..." : "Войти"}
      </Button>
      <p className="text-center text-sm text-gray-500">
        Нет аккаунта?{" "}
        <Link className="font-semibold text-gray-900 hover:underline" href="/register">
          Регистрация
        </Link>
      </p>
    </form>
  );
}
