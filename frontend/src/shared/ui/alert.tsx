import * as React from "react";

import { cn } from "@/shared/lib/cn";

type AlertVariant = "error" | "info" | "success";

const variants: Record<AlertVariant, string> = {
  error: "border-red-100 bg-red-50 text-red-600",
  info: "border-gray-200 bg-gray-50 text-gray-600",
  success: "border-gray-200 bg-gray-900 text-white"
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
