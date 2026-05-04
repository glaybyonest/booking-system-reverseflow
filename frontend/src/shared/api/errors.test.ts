import { describe, expect, it } from "vitest";

import { AppApiError, friendlyApiError } from "@/shared/api/errors";

describe("friendlyApiError", () => {
  it("maps backend seat conflicts to Russian text", () => {
    const error = new AppApiError({
      code: "SEAT_NOT_AVAILABLE",
      message: "Seat is not available",
      status: 409
    });

    expect(friendlyApiError(error)).toContain("место");
  });
});
