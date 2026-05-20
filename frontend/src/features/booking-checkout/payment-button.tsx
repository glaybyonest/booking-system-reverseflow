"use client";

import { CreditCard } from "lucide-react";

import { Button } from "@/shared/ui/button";

export function PaymentButton({
  pending,
  disabled,
  onClick,
  forceStatus = "succeeded"
}: {
  pending: boolean;
  disabled: boolean;
  forceStatus?: "succeeded" | "failed";
  onClick(forceStatus: "succeeded" | "failed"): void;
}) {
  return (
    <Button className="w-full" size="lg" disabled={disabled || pending} onClick={() => onClick(forceStatus)}>
      <CreditCard className="h-4 w-4" />
      {pending
        ? "Обрабатываем..."
        : forceStatus === "failed"
          ? "Проверить сценарий отказа"
          : "Оплатить демо-картой"}
    </Button>
  );
}
