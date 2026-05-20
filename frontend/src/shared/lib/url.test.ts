import { describe, expect, it } from "vitest";

import {
  buildEventPath,
  parseEventRouteId,
  slugifySegment,
  toSafeExternalUrl,
  toSafeRedirectPath
} from "@/shared/lib/url";

describe("url helpers", () => {
  it("builds event paths with a readable slug", () => {
    expect(buildEventPath("event-1", "Ночной концерт в Москве")).toBe(
      "/events/%D0%BD%D0%BE%D1%87%D0%BD%D0%BE%D0%B8-%D0%BA%D0%BE%D0%BD%D1%86%D0%B5%D1%80%D1%82-%D0%B2-%D0%BC%D0%BE%D1%81%D0%BA%D0%B2%D0%B5--event-1"
    );
  });

  it("extracts the event id from a slug route", () => {
    expect(parseEventRouteId("ночнои-концерт-в-москве--event-1")).toBe("event-1");
    expect(parseEventRouteId("event-1")).toBe("event-1");
  });

  it("sanitizes external URLs", () => {
    expect(toSafeExternalUrl("https://reserveflow.app/events")).toBe(
      "https://reserveflow.app/events"
    );
    expect(toSafeExternalUrl("javascript:alert(1)")).toBeNull();
    expect(toSafeExternalUrl("data:text/html,<script>alert(1)</script>")).toBeNull();
  });

  it("keeps redirects inside the app", () => {
    expect(toSafeRedirectPath("/events/ночнои-концерт--event-1?from=login")).toBe(
      "/events/%D0%BD%D0%BE%D1%87%D0%BD%D0%BE%D0%B8-%D0%BA%D0%BE%D0%BD%D1%86%D0%B5%D1%80%D1%82--event-1?from=login"
    );
    expect(toSafeRedirectPath("javascript:alert(1)")).toBe("/events");
    expect(toSafeRedirectPath("https://evil.example/steal")).toBe("/events");
  });

  it("falls back to a default slug when the title is empty", () => {
    expect(slugifySegment("")).toBe("event");
  });
});
