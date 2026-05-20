import * as React from "react";

import { cn } from "@/shared/lib/cn";

type AlertVariant = "error" | "info" | "success" | "warning";

const variants: Record<AlertVariant, string> = {
  error: "border-err/20 bg-err-soft text-err-fg",
  info: "border-info/20 bg-info-soft text-info-fg",
  success: "border-ok/20 bg-ok-soft text-ok-fg",
  warning: "border-warn/20 bg-warn-soft text-warn-fg"
};

export function Alert({
  className,
  variant = "info",
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { variant?: AlertVariant }) {
  return (
    <div
      className={cn("rounded-xl border p-3 text-sm leading-relaxed", variants[variant], className)}
      {...props}
    />
  );
}
