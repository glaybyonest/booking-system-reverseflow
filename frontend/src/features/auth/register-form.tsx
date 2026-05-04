"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
      <label className="block space-y-2 text-sm font-medium">
        <span>Имя</span>
        <Input autoComplete="name" {...form.register("name")} />
        {form.formState.errors.name ? (
          <span className="text-xs text-red-600">{form.formState.errors.name.message}</span>
        ) : null}
      </label>
      <label className="block space-y-2 text-sm font-medium">
        <span>Email</span>
        <Input type="email" autoComplete="email" {...form.register("email")} />
        {form.formState.errors.email ? (
          <span className="text-xs text-red-600">{form.formState.errors.email.message}</span>
        ) : null}
      </label>
      <label className="block space-y-2 text-sm font-medium">
        <span>Пароль</span>
        <Input type="password" autoComplete="new-password" {...form.register("password")} />
        {form.formState.errors.password ? (
          <span className="text-xs text-red-600">{form.formState.errors.password.message}</span>
        ) : null}
      </label>
      <label className="block space-y-2 text-sm font-medium">
        <span>Повторите пароль</span>
        <Input type="password" autoComplete="new-password" {...form.register("confirmPassword")} />
        {form.formState.errors.confirmPassword ? (
          <span className="text-xs text-red-600">
            {form.formState.errors.confirmPassword.message}
          </span>
        ) : null}
      </label>
      <Button className="w-full" size="lg" type="submit" disabled={register.isPending}>
        {register.isPending ? "Создаем аккаунт..." : "Зарегистрироваться"}
      </Button>
      <p className="text-center text-sm text-gray-500">
        Уже есть аккаунт?{" "}
        <Link className="font-semibold text-gray-900 hover:underline" href="/login">
          Войти
        </Link>
      </p>
    </form>
  );
}
