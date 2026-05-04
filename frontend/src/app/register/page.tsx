import { Card, CardContent } from "@/shared/ui/card";
import { RegisterForm } from "@/features/auth/register-form";
import { Logo } from "@/widgets/header/header";

export default function RegisterPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-[#F8F9FA] px-4 py-10">
      <Card className="w-full max-w-md">
        <CardContent>
          <div className="mb-8 flex flex-col items-center text-center">
            <Logo />
            <h1 className="mt-8 text-2xl font-bold">Регистрация</h1>
            <p className="mt-2 text-sm text-gray-500">Создайте аккаунт и сразу переходите к выбору мест.</p>
          </div>
          <RegisterForm />
        </CardContent>
      </Card>
    </main>
  );
}
