import { describe, expect, it } from "vitest";

import { normalizeEvent } from "@/shared/api/mappers";

describe("normalizeEvent", () => {
  it("drops unsafe imported urls", () => {
    const event = normalizeEvent({
      id: "event-1",
      title: "Unsafe event",
      sourceUrl: "javascript:alert(1)",
      posterUrl: "data:text/html,<script>alert(1)</script>",
      externalLinks: [
        {
          id: "link-1",
          externalSource: "manual",
          externalId: "1",
          sourceUrl: "javascript:alert(1)"
        }
      ]
    });

    expect(event.sourceUrl).toBeUndefined();
    expect(event.posterUrl).toBeUndefined();
    expect(event.externalLinks?.[0]?.sourceUrl).toBeUndefined();
  });
});
