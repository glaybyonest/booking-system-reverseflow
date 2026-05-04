import { Suspense } from "react";

import { Card, CardContent } from "@/shared/ui/card";
import { LoginForm } from "@/features/auth/login-form";
import { Logo } from "@/widgets/header/header";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-[#F8F9FA] px-4 py-10">
      <Card className="w-full max-w-md">
        <CardContent>
          <div className="mb-8 flex flex-col items-center text-center">
            <Logo />
            <h1 className="mt-8 text-2xl font-bold">Войти в ReserveFlow</h1>
            <p className="mt-2 text-sm text-gray-500">Продолжите бронирование без хранения токенов в браузере.</p>
          </div>
          <Suspense>
            <LoginForm />
          </Suspense>
        </CardContent>
      </Card>
    </main>
  );
}
