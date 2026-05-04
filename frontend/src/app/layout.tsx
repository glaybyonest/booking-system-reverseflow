import type { Metadata } from "next";

import "@fontsource-variable/inter";
import "@/app/globals.css";
import { Providers } from "@/app/providers";

export const metadata: Metadata = {
  title: "ReserveFlow",
  description: "Бронирование мест на мероприятия быстро и безопасно."
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body className="min-h-screen bg-[#F8F9FA] font-sans text-gray-900 antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
