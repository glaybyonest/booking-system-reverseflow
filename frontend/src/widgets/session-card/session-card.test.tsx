import { render, screen } from "@testing-library/react";
import React from "react";
import { describe, expect, it, vi } from "vitest";

import type { Session } from "@/entities/session/types";
import { SessionCard } from "@/widgets/session-card/session-card";

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

const baseSession: Session = {
  id: "session-1",
  eventId: "event-1",
  hallId: undefined,
  startsAt: "2030-01-01T18:00:00+03:00",
  endsAt: "2030-01-01T20:00:00+03:00",
  status: "scheduled",
  isBookable: false,
  hallName: "Demo Hall"
};

describe("SessionCard", () => {
  it("hides booking button for unavailable session", () => {
    render(<SessionCard session={baseSession} />);

    expect(screen.queryByText("Купить билеты")).not.toBeInTheDocument();
    expect(screen.getByText("Бронирование недоступно")).toBeInTheDocument();
  });

  it("shows booking button for bookable session", () => {
    render(<SessionCard session={{ ...baseSession, id: "session-2", isBookable: true }} />);

    expect(screen.getByText("Купить билеты")).toBeInTheDocument();
  });
});
