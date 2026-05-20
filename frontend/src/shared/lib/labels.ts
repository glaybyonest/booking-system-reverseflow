import type { SeatStatus } from "@/entities/seat/types";

export const eventStatusLabels: Record<string, string> = {
  published: "Опубликовано",
  draft: "Черновик",
  archived: "В архиве"
};

export const sessionStatusLabels: Record<string, string> = {
  scheduled: "По расписанию",
  cancelled: "Отменен"
};

export const seatStatusLabels: Record<SeatStatus, string> = {
  available: "свободно",
  held: "удерживается",
  booked: "занято"
};

export const sourceLabels: Record<string, string> = {
  manual: "ReserveFlow",
  reserveflow: "ReserveFlow",
  yandex_afisha: "Яндекс Афиша"
};

export const bookingModeLabels: Record<string, string> = {
  reserveflow_managed: "Можно забронировать",
  general_admission: "General admission",
  bookable: "Можно забронировать"
};

export function displayLabel(value: string | undefined, labels: Record<string, string>) {
  if (!value) return "Уточняется";
  return labels[value] ?? value;
}
