import Link from "next/link";

import { Logo } from "@/widgets/header/header";

const SERVICE_LINKS = ["Мероприятия", "Площадки", "Скидки", "Подарочные билеты"];
const COMPANY_LINKS = ["О нас", "Карьера", "Партнёрам", "Контакты"];
const HELP_LINKS = ["Как забронировать", "Возврат", "Политика", "Поддержка"];

export function Footer() {
  return (
    <footer className="mt-auto border-t border-border bg-white">
      <div className="mx-auto max-w-[1440px] px-6 py-10">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div>
            <Logo size="sm" />
            <p className="mt-3 max-w-[180px] text-xs leading-relaxed text-mute">
              Бронирование мест на мероприятия с прозрачным удержанием и понятными статусами.
              Прототип MVP, 2026.
            </p>
          </div>
          <FooterCol title="СЕРВИС" links={SERVICE_LINKS} />
          <FooterCol title="КОМПАНИЯ" links={COMPANY_LINKS} />
          <FooterCol title="ПОМОЩЬ" links={HELP_LINKS} />
        </div>
        <div className="mt-8 flex items-center justify-between border-t border-border pt-6">
          <p className="text-xs text-mute-2">© 2026 ReserveFlow. Все права защищены.</p>
          <span className="text-xs text-mute-2">Версия 0.4 · MVP</span>
        </div>
      </div>
    </footer>
  );
}

function FooterCol({ title, links }: { title: string; links: string[] }) {
  return (
    <div>
      <p className="mb-3 text-[10px] font-bold uppercase tracking-widest text-mute-2">{title}</p>
      <ul className="space-y-2">
        {links.map((link) => (
          <li key={link}>
            <Link
              href="#"
              className="text-sm text-mute transition-colors hover:text-ink"
            >
              {link}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
