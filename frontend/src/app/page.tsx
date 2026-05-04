import { CalendarSearch, Clock3, LayoutGrid, ShieldCheck } from "lucide-react";
import Link from "next/link";

import { routes } from "@/shared/lib/routes";
import { Button } from "@/shared/ui/button";
import { Header } from "@/widgets/header/header";

const advantages = [
  {
    title: "Выбор мероприятия",
    description: "Удобный каталог событий с понятными карточками и быстрым переходом к сеансам.",
    icon: CalendarSearch
  },
  {
    title: "Интерактивная схема",
    description: "Актуальная карта зала показывает свободные, удержанные и занятые места.",
    icon: LayoutGrid
  },
  {
    title: "Удержание места",
    description: "Выбранное место закрепляется за вами на 10 минут для спокойной оплаты.",
    icon: Clock3,
    highlighted: true
  },
  {
    title: "Безопасная бронь",
    description: "Backend защищает места от double booking через транзакции PostgreSQL.",
    icon: ShieldCheck
  }
];

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col bg-[#F8F9FA]">
      <Header variant="public" />
      <main className="mt-20 flex flex-1 flex-col items-center justify-center px-6 text-center">
        <h1 className="max-w-4xl text-5xl font-extrabold leading-[1.1] tracking-tight md:text-7xl">
          Бронируйте места на мероприятия{" "}
          <span className="bg-gradient-to-r from-gray-900 to-gray-500 bg-clip-text text-transparent">
            быстро и безопасно.
          </span>
        </h1>
        <p className="mb-10 mt-6 max-w-2xl text-lg text-gray-500 md:text-xl">
          Выбирайте событие, смотрите схему зала, удерживайте место и оплачивайте mock payment
          без риска двойной брони.
        </p>
        <Link href={routes.events}>
          <Button className="rounded-full px-8 py-4 text-lg shadow-xl hover:-translate-y-0.5 hover:shadow-2xl">
            Смотреть мероприятия
          </Button>
        </Link>
      </main>
      <section className="mx-auto w-full max-w-7xl px-6 pb-24 pt-24">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
          {advantages.map((item) => {
            const Icon = item.icon;
            return (
              <div
                key={item.title}
                className={
                  item.highlighted
                    ? "rounded-[2rem] bg-gray-900 p-8 text-white shadow-xl transition-transform hover:-translate-y-1"
                    : "rounded-[2rem] border border-gray-100 bg-white p-8 shadow-sm transition-shadow hover:shadow-md"
                }
              >
                <div
                  className={
                    item.highlighted
                      ? "mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-800"
                      : "mb-6 flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-50"
                  }
                >
                  <Icon className="h-6 w-6" />
                </div>
                <h3 className="mb-2 text-xl font-bold">{item.title}</h3>
                <p className={item.highlighted ? "text-sm leading-relaxed text-gray-300" : "text-sm leading-relaxed text-gray-500"}>
                  {item.description}
                </p>
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
