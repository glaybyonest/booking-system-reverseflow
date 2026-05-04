import * as React from "react";

import { cn } from "@/shared/lib/cn";

type BadgeVariant =
  | "default"
  | "dark"
  | "success"
  | "muted"
  | "danger"
  | "warning"
  | "outline";

const variants: Record<BadgeVariant, string> = {
  default: "bg-gray-100 text-gray-600",
  dark: "bg-gray-900 text-white",
  success: "bg-gray-900 text-white",
  muted: "bg-gray-100 text-gray-400",
  danger: "border border-red-100 bg-red-50 text-red-600",
  warning: "border border-orange-200 bg-orange-50 text-orange-700",
  outline: "border border-gray-200 bg-white text-gray-600"
};

export function Badge({
  className,
  variant = "default",
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { variant?: BadgeVariant }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold",
        variants[variant],
        className
      )}
      {...props}
    />
  );
}
