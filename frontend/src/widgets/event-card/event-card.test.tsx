import { render, screen } from "@testing-library/react";
import React from "react";
import { describe, expect, it, vi } from "vitest";

import type { Event } from "@/entities/event/types";
import { EventCard } from "@/widgets/event-card/event-card";

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: React.PropsWithChildren<{ href: string } & React.AnchorHTMLAttributes<HTMLAnchorElement>>) => (
    <a href={href} {...props}>
      {children}
    </a>
  )
}));

const baseEvent: Event = {
  id: "event-1",
  title: "Ночной концерт в Москве",
  description: "Живое выступление в центре города.",
  category: "concert",
  status: "published",
  source: "kudago",
  externalSource: "kudago",
  sourceUrl: "https://kudago.com/event-1",
  bookingMode: "external_link_only",
  startsAt: "2030-01-01T18:00:00+03:00",
  endsAt: "2030-01-01T20:00:00+03:00",
  isImported: true,
  venue: {
    id: "venue-1",
    name: "Дом музыки",
    address: "Москва, Космодамианская набережная, 52",
    city: "Москва",
    latitude: 55.7343,
    longitude: 37.6467
  }
};

describe("EventCard", () => {
  it("renders source badge and venue details for imported event", () => {
    render(<EventCard event={baseEvent} />);

    expect(screen.getByText("Ночной концерт в Москве")).toBeInTheDocument();
    expect(screen.getByText("KudaGo")).toBeInTheDocument();
    expect(screen.getByText("Дом музыки")).toBeInTheDocument();
    expect(screen.getByText("Открыть у организатора")).toBeInTheDocument();
  });

  it("uses a readable slug in internal event links", () => {
    render(<EventCard event={{ ...baseEvent, bookingMode: "reserveflow_managed" }} />);

    expect(screen.getByText("Подробнее")).toHaveAttribute(
      "href",
      "/events/%D0%BD%D0%BE%D1%87%D0%BD%D0%BE%D0%B8-%D0%BA%D0%BE%D0%BD%D1%86%D0%B5%D1%80%D1%82-%D0%B2-%D0%BC%D0%BE%D1%81%D0%BA%D0%B2%D0%B5--event-1"
    );
  });
});
