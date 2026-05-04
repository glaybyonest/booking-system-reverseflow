import { ArrowRight } from "lucide-react";

import { Button } from "@/shared/ui/button";

export function HoldSeatButton({
  disabled,
  pending
}: {
  disabled: boolean;
  pending: boolean;
}) {
  return (
    <Button className="w-full" size="lg" type="submit" disabled={disabled || pending}>
      {pending ? "Удерживаем..." : "Удержать место"}
      <ArrowRight className="h-4 w-4" />
    </Button>
  );
}
