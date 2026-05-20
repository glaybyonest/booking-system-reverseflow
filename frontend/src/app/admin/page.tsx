import Link from "next/link";

import { Header } from "@/widgets/header/header";

export default function AdminHomePage() {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auto" />
      <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6">
        <div className="rounded-[2rem] border border-gray-200 bg-white p-8 shadow-sm">
          <h1 className="text-3xl font-bold">Admin layouts</h1>
          <p className="mt-3 max-w-3xl text-sm text-gray-500">
            Откройте карточку нужного события и используйте кнопки `Session layout` или `Hall fallback`
            рядом с сеансом, чтобы настроить точную схему мест для KudaGo-события.
          </p>
          <div className="mt-6">
            <Link
              href="/events"
              className="inline-flex items-center justify-center rounded-2xl bg-gray-900 px-5 py-3 text-sm font-medium text-white shadow-md transition-all hover:bg-gray-800 hover:shadow-lg"
            >
              Перейти в каталог событий
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
