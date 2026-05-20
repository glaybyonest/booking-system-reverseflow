import { RegisterForm } from "@/features/auth/register-form";
import { Logo } from "@/widgets/header/header";

export default function RegisterPage() {
  return (
    <main className="flex min-h-screen bg-bg">
      {/* ── Left: form ── */}
      <div className="flex flex-1 flex-col items-center justify-center px-8 py-12">
        <div className="w-full max-w-sm">
          <div className="mb-8">
            <Logo />
          </div>
          <h1 className="text-2xl font-extrabold text-ink">Регистрация</h1>
          <p className="mt-2 text-sm text-mute">
            Создайте аккаунт и переходите к выбору мест.
          </p>
          <div className="mt-8">
            <RegisterForm />
          </div>
        </div>
      </div>

      {/* ── Right: feature highlights ── */}
      <div className="hidden flex-col justify-center bg-ink px-12 py-16 lg:flex lg:w-[420px] xl:w-[480px]">
        <div className="mb-8">
          <Logo light />
        </div>
        <h2 className="text-3xl font-extrabold leading-snug text-white">
          Бронируйте места на мероприятия без риска двойной брони.
        </h2>
        <p className="mt-4 text-sm leading-relaxed text-white/60">
          Регистрация занимает меньше минуты. После этого вы сможете удерживать места на 10 минут и оплачивать в любое удобное время.
        </p>
        <div className="mt-10 grid grid-cols-1 gap-4">
          {[
            { title: "Безопасный hold", desc: "Место закреплено за вами" },
            { title: "10-минутный таймер", desc: "Успейте оплатить спокойно" },
            { title: "Понятные статусы", desc: "Ожидает, подтверждена, истекла" }
          ].map((item) => (
            <div key={item.title} className="rounded-xl border border-white/10 bg-white/5 p-4">
              <p className="text-sm font-bold text-white">{item.title}</p>
              <p className="mt-0.5 text-xs text-white/50">{item.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}
