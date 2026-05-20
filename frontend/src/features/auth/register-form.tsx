"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Lock, Mail, User } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";

import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { RegisterInput, registerSchema } from "@/features/auth/auth.schemas";
import { useRegister } from "@/features/auth/auth.hooks";

export function RegisterForm() {
  const router = useRouter();
  const register = useRegister();
  const form = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: { name: "", email: "", password: "", confirmPassword: "" }
  });

  const onSubmit = form.handleSubmit((values) => {
    register.mutate(
      { name: values.name, email: values.email, password: values.password },
      {
        onSuccess: (result) => router.push(result.authenticated ? "/events" : "/login")
      }
    );
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      {register.error ? <Alert variant="error">{friendlyApiError(register.error)}</Alert> : null}

      <div className="space-y-1.5">
        <label className="text-sm font-medium text-ink">Имя</label>
        <div className="relative">
          <User className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-mute-2" />
          <Input
            autoComplete="name"
            placeholder="Иван Иванов"
            className="pl-10"
            {...form.register("name")}
          />
        </div>
        {form.formState.errors.name ? (
          <p className="text-xs text-err">{form.formState.errors.name.message}</p>
        ) : null}
      </div>

      <div className="space-y-1.5">
        <label className="text-sm font-medium text-ink">Email</label>
        <div className="relative">
          <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-mute-2" />
          <Input
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            className="pl-10"
            {...form.register("email")}
          />
        </div>
        {form.formState.errors.email ? (
          <p className="text-xs text-err">{form.formState.errors.email.message}</p>
        ) : null}
      </div>

      <div className="space-y-1.5">
        <label className="text-sm font-medium text-ink">Пароль</label>
        <div className="relative">
          <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-mute-2" />
          <Input
            type="password"
            autoComplete="new-password"
            placeholder="••••••••"
            className="pl-10"
            {...form.register("password")}
          />
        </div>
        {form.formState.errors.password ? (
          <p className="text-xs text-err">{form.formState.errors.password.message}</p>
        ) : null}
      </div>

      <div className="space-y-1.5">
        <label className="text-sm font-medium text-ink">Повторите пароль</label>
        <div className="relative">
          <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-mute-2" />
          <Input
            type="password"
            autoComplete="new-password"
            placeholder="••••••••"
            className="pl-10"
            {...form.register("confirmPassword")}
          />
        </div>
        {form.formState.errors.confirmPassword ? (
          <p className="text-xs text-err">{form.formState.errors.confirmPassword.message}</p>
        ) : null}
      </div>

      <Button className="w-full" size="lg" type="submit" disabled={register.isPending}>
        {register.isPending ? "Создаём аккаунт..." : "Зарегистрироваться →"}
      </Button>

      <p className="text-center text-sm text-mute">
        Уже есть аккаунт?{" "}
        <Link className="font-semibold text-ink hover:underline" href="/login">
          Войти
        </Link>
      </p>
    </form>
  );
}
